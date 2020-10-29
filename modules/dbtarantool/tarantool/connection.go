package tarantool

import (
	"bufio"
	"errors"
	"io"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultConnectTimeout = time.Second
)

var (
	ErrEmptyDefaultSpace = errors.New("zero-length default space or unnecessary slash in dsn.path")
	ErrSyncFailed        = errors.New("SYNC failed")
)

type Options struct {
	ConnectTimeout time.Duration
	QueryTimeout   time.Duration
	DefaultSpace   string
	User           string
	Password       string
	UUID           string
	ReplicaSetUUID string
}

type Greeting struct {
	Version []byte
	Auth    []byte
}

type Connection struct {
	requestID uint32
	requests  *requestMap
	writeChan chan *packedPacket // packed messages with header
	closeOnce sync.Once
	exit      chan bool
	closed    chan bool
	conn      net.Conn
	// options
	queryTimeout   time.Duration
	Greeting       *Greeting
	packData       *packData
	remoteAddr     string
	firstError     error
	firstErrorLock *sync.Mutex
}

// Connect to tarantool instance with options.
// Returned Connection could be used to execute queries.
func Connect(scheme string, host string, options *Options) (conn *Connection, err error) {
	var opts Options
	var deadline time.Time

	opts = parseOptions(options)
	if conn, err = newConn(scheme, host, opts); err != nil {
		return
	}
	// set schema pulling deadline
	deadline = time.Now().Add(opts.ConnectTimeout)
	conn.conn.SetDeadline(deadline)
	if err = conn.pullSchema(); err != nil {
		conn.conn.Close()
		conn = nil
		return
	}
	// remove deadline
	conn.conn.SetDeadline(time.Time{})
	go conn.worker(conn.conn)

	return
}

func newConn(network, addr string, opts Options) (conn *Connection, err error) {
	defer func() { // close opened connection if error
		if err != nil && conn != nil {
			if conn.conn != nil {
				conn.conn.Close()
			}
			conn = nil
		}
	}()

	conn = &Connection{
		remoteAddr:     addr,
		requests:       newRequestMap(),
		writeChan:      make(chan *packedPacket, 256),
		exit:           make(chan bool),
		closed:         make(chan bool),
		firstErrorLock: &sync.Mutex{},
		packData:       newPackData(opts.DefaultSpace),
		queryTimeout:   opts.QueryTimeout,
	}

	conn.conn, err = net.DialTimeout(network, conn.remoteAddr, opts.ConnectTimeout)
	if err != nil {
		return
	}

	greeting := make([]byte, 128)

	connectDeadline := time.Now().Add(opts.ConnectTimeout)
	conn.conn.SetDeadline(connectDeadline)
	// removing deadline deferred
	defer conn.conn.SetDeadline(time.Time{})

	_, err = io.ReadFull(conn.conn, greeting)
	if err != nil {
		return
	}

	conn.Greeting = &Greeting{
		Version: greeting[:64],
		Auth:    greeting[64:108],
	}

	// try to authenticate if user have been provided
	if len(opts.User) > 0 {
		var authResponse *Packet

		requestID := conn.nextID()

		pp := packIproto(0, requestID)
		defer pp.Release()

		pp.code, err = (&Auth{
			User:         opts.User,
			Password:     opts.Password,
			GreetingAuth: conn.Greeting.Auth,
		}).Pack(conn.packData, &pp.buffer)
		if err != nil {
			return
		}

		if _, err = pp.WriteTo(conn.conn); err != nil {
			return
		}

		if authResponse, err = readPacket(conn.conn); err != nil {
			return
		}

		if authResponse.requestID != requestID {
			err = ErrSyncFailed
			return
		}

		if authResponse.Result != nil && authResponse.Result.Error != nil {
			err = authResponse.Result.Error
			return
		}
	}

	return
}

func parseOptions(options *Options) (opts Options) {
	if options != nil {
		opts = *options
	}
	if opts.ConnectTimeout.Nanoseconds() == 0 {
		opts.ConnectTimeout = defaultConnectTimeout
	}

	return
}

func (conn *Connection) pullSchema() (err error) {
	// select space and index schema
	request := func(q Query) (*Result, error) {
		var err error

		requestID := conn.nextID()

		pp := packIproto(0, requestID)
		defer pp.Release()

		pp.code, err = q.Pack(conn.packData, &pp.buffer)
		if err != nil {
			return nil, err
		}

		if _, err = pp.WriteTo(conn.conn); err != nil {
			return nil, err
		}

		response, err := readPacket(conn.conn)
		if err != nil {
			return nil, err
		}

		if response.requestID != requestID {
			return nil, errors.New("Bad response requestID")
		}

		if response.Result == nil {
			return nil, errors.New("Nil response result")
		}

		if response.Result.Error != nil {
			return nil, response.Result.Error
		}

		return response.Result, nil
	}

	res, err := request(&Select{
		Space:    ViewSpace,
		Key:      0,
		Iterator: IterAll,
	})
	if err != nil {
		return
	}

	for _, space := range res.Data {
		conn.packData.spaceMap[space[2].(string)] = space[0].(uint16)
	}

	res, err = request(&Select{
		Space:    ViewIndex,
		Key:      0,
		Iterator: IterAll,
	})
	if err != nil {
		return
	}

	for _, index := range res.Data {
		spaceID := index[0].(uint16)
		indexID := index[1].(int8)
		indexName := index[2].(string)
		indexAttr := index[4].(map[string]interface{}) // e.g: {"unique": true}
		indexFields := index[5].([]interface{})        // e.g: [[0 num] [1 str]]

		indexSpaceMap, exists := conn.packData.indexMap[spaceID]
		if !exists {
			indexSpaceMap = make(map[string]uint64)
			conn.packData.indexMap[spaceID] = indexSpaceMap
		}
		indexSpaceMap[indexName] = uint64(indexID)

		// build list of primary key field numbers for this space, if the PK is detected
		if indexAttr != nil && indexID == 0 {
			if unique, ok := indexAttr["unique"]; ok && unique.(bool) {
				pk := make([]int, len(indexFields))
				for i := range indexFields {
					f := indexFields[i].([]interface{})
					pk[i] = int(f[0].(int8))
				}
				conn.packData.primaryKeyMap[spaceID] = pk
			}
		}
	}

	return
}

func (conn *Connection) nextID() uint32 {
	return atomic.AddUint32(&conn.requestID, 1)
}

func (conn *Connection) stop() {
	conn.closeOnce.Do(func() {
		// debug.PrintStack()
		close(conn.exit)
		conn.conn.Close()
		runtime.GC()
	})
}

func (conn *Connection) GetPrimaryKeyFields(space interface{}) ([]int, bool) {
	if conn.packData == nil {
		return nil, false
	}

	var spaceID uint16
	switch space := space.(type) {
	case int:
		spaceID = uint16(space)
	case uint:
		spaceID = uint16(space)
	case uint32:
		spaceID = uint16(space)
	case uint64:
		spaceID = uint16(space)
	case string:
		spaceID = conn.packData.spaceMap[space]
	default:
		return nil, false
	}

	f, ok := conn.packData.primaryKeyMap[spaceID]
	return f, ok
}

func (conn *Connection) Close() {
	conn.stop()
	<-conn.closed
}

func (conn *Connection) String() string {
	return conn.remoteAddr
}

func (conn *Connection) IsClosed() (bool, error) {
	select {
	case <-conn.exit:
		return true, conn.getError()
	default:
		return false, conn.getError()
	}
}

func (conn *Connection) getError() error {
	conn.firstErrorLock.Lock()
	defer conn.firstErrorLock.Unlock()
	return conn.firstError
}

func (conn *Connection) setError(err error) {
	if err != nil && err != io.EOF {
		conn.firstErrorLock.Lock()
		if conn.firstError == nil {
			conn.firstError = err
		}
		conn.firstErrorLock.Unlock()
	}
}

func (conn *Connection) worker(tcpConn net.Conn) {
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		err := writer(tcpConn, conn.writeChan, conn.exit)
		conn.setError(err)
		conn.stop()
		wg.Done()
	}()

	go func() {
		err := conn.reader(tcpConn)
		conn.setError(err)
		conn.stop()
		wg.Done()
	}()

	wg.Wait()

	// release all pending packets
	writeChan := conn.writeChan
	conn.writeChan = nil

CLEANUP_LOOP:
	for {
		select {
		case pp := <-writeChan:
			pp.Release()
		default:
			break CLEANUP_LOOP
		}
	}

	// send error reply to all pending requests
	conn.requests.CleanUp(func(req *request) {
		req.replyChan <- &Result{
			Error: ConnectionClosedError(conn),
		}
		close(req.replyChan)
	})

	close(conn.closed)
}

func writer(tcpConn io.Writer, writeChan chan *packedPacket, stopChan chan bool) (err error) {
	w := bufio.NewWriter(tcpConn)

WRITER_LOOP:
	for {
		select {
		case packet, ok := <-writeChan:
			if !ok {
				break WRITER_LOOP
			}

			_, err = packet.WriteTo(w)
			if err != nil {
				break WRITER_LOOP
			}
		case <-stopChan:
			break WRITER_LOOP
		default:
			if err = w.Flush(); err != nil {
				break WRITER_LOOP
			}

			// same without flush
			select {
			case packet, ok := <-writeChan:
				if !ok {
					break WRITER_LOOP
				}

				_, err = packet.WriteTo(w)
				if err != nil {
					break WRITER_LOOP
				}
			case <-stopChan:
				break WRITER_LOOP
			}
		}
	}

	return
}

func (conn *Connection) reader(tcpConn io.Reader) (err error) {
	var packet *Packet
	var pp *packedPacket
	var req *request

	r := bufio.NewReaderSize(tcpConn, DefaultReaderBufSize)

READER_LOOP:
	for {
		// read raw bytes
		pp, err = readPacked(r)
		if err != nil {
			break READER_LOOP
		}

		packet, err = decodePacket(pp)
		if err != nil {
			break READER_LOOP
		}

		req = conn.requests.Pop(packet.requestID)
		if req != nil {
			res := &Result{}
			if packet.Result != nil {
				res = packet.Result
			}
			req.replyChan <- res
			close(req.replyChan)
		}

		pp.Release()
		pp = nil
	}

	if pp != nil {
		pp.Release()
	}
	return
}

func readPacket(r io.Reader) (p *Packet, err error) {
	pp, err := readPacked(r)
	if err != nil {
		return nil, err
	}
	defer pp.Release()

	p, err = decodePacket(pp)
	if err != nil {
		return nil, err
	}

	return p, nil
}
