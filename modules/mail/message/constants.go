package message // import "github.com/webnice/kit/v1/modules/mail/message"

const (
	// DeafultContentType default embeded Content-Type
	DeafultContentType = `text/plain`

	// DeafultAttachContentType default attach Content-Type
	DeafultAttachContentType = `application/octet-stream`

	// HeaderMimeVersion header Mime-Version
	HeaderMimeVersion Header = `Mime-Version`

	// HeaderDate header Date
	HeaderDate Header = `Date`

	// HeaderFrom header From
	HeaderFrom Header = `From`

	// HeaderTo header to
	HeaderTo Header = `To`

	// HeaderToCC header carbon copy
	HeaderToCC Header = `Cc`

	// HeaderToBCC header to blind carbon copy
	HeaderToBCC Header = `Bcc`

	// HeaderSubject header subject
	HeaderSubject Header = `Subject`
)
