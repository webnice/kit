package kong

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Path records the nodes and parsed values from the current command-line.
type Path struct {
	Parent *Node

	// One of these will be non-nil.
	App        *Application
	Positional *Positional
	Flag       *Flag
	Argument   *Argument
	Command    *Command

	// Flags added by this node.
	Flags []*Flag

	// True if this Path element was created as the result of a resolver.
	Resolved bool
}

// Node returns the Node associated with this Path, or nil if Path is a non-Node.
func (p *Path) Node() *Node {
	switch {
	case p.App != nil:
		return p.App.Node

	case p.Argument != nil:
		return p.Argument

	case p.Command != nil:
		return p.Command
	}
	return nil
}

// Visitable returns the Visitable for this path element.
func (p *Path) Visitable() Visitable {
	switch {
	case p.App != nil:
		return p.App

	case p.Argument != nil:
		return p.Argument

	case p.Command != nil:
		return p.Command

	case p.Flag != nil:
		return p.Flag

	case p.Positional != nil:
		return p.Positional
	}
	return nil
}

// Context contains the current parse context.
type Context struct {
	*Kong
	// A trace through parsed nodes.
	Path []*Path
	// Original command-line arguments.
	Args []string
	// Error that occurred during trace, if any.
	Error error

	values    map[*Value]reflect.Value // Temporary values during tracing.
	bindings  bindings
	resolvers []Resolver // Extra context-specific resolvers.
	scan      *Scanner
}

// Trace path of "args" through the grammar tree.
//
// The returned Context will include a Path of all commands, arguments, positionals and flags.
//
// This just constructs a new trace. To fully apply the trace you must call Reset(), Resolve(),
// Validate() and Apply().
func Trace(k *Kong, args []string) (*Context, error) {
	c := &Context{
		Kong: k,
		Args: args,
		Path: []*Path{
			{App: k.Model, Flags: k.Model.Flags},
		},
		values:   map[*Value]reflect.Value{},
		scan:     Scan(args...),
		bindings: bindings{},
	}
	c.Error = c.trace(c.Model.Node)
	return c, nil
}

// Bind adds bindings to the Context.
func (c *Context) Bind(args ...any) {
	c.bindings.add(args...)
}

// BindTo adds a binding to the Context.
//
// This will typically have to be called like so:
//
//	BindTo(impl, (*MyInterface)(nil))
func (c *Context) BindTo(impl, iface any) {
	c.bindings.addTo(impl, iface)
}

// BindToProvider allows binding of provider functions.
//
// This is useful when the Run() function of different commands require different values that may
// not all be initialisable from the main() function.
func (c *Context) BindToProvider(provider any) error {
	return c.bindings.addProvider(provider)
}

// Value returns the value for a particular path element.
func (c *Context) Value(path *Path) reflect.Value {
	switch {
	case path.Positional != nil:
		return c.values[path.Positional]
	case path.Flag != nil:
		return c.values[path.Flag.Value]
	case path.Argument != nil:
		return c.values[path.Argument.Argument]
	}
	panic("can only retrieve value for flag, argument or positional")
}

// Selected command or argument.
func (c *Context) Selected() *Node {
	var selected *Node
	for _, path := range c.Path {
		switch {
		case path.Command != nil:
			selected = path.Command
		case path.Argument != nil:
			selected = path.Argument
		}
	}
	return selected
}

// Empty returns true if there were no arguments provided.
func (c *Context) Empty() bool {
	for _, path := range c.Path {
		if !path.Resolved && path.App == nil {
			return false
		}
	}
	return true
}

// Validate the current context.
func (c *Context) Validate() error { // nolint: gocyclo
	err := Visit(c.Model, func(node Visitable, next Next) error {
		switch node := node.(type) {
		case *Value:
			_, ok := os.LookupEnv(node.Tag.Env)
			if node.Enum != "" && (!node.Required || node.HasDefault || (node.Tag.Env != "" && ok)) {
				if err := checkEnum(node, node.Target); err != nil {
					return err
				}
			}

		case *Flag:
			_, ok := os.LookupEnv(node.Tag.Env)
			if node.Enum != "" && (!node.Required || node.HasDefault || (node.Tag.Env != "" && ok)) {
				if err := checkEnum(node.Value, node.Target); err != nil {
					return err
				}
			}
		}
		return next(nil)
	})
	if err != nil {
		return err
	}
	for _, el := range c.Path {
		var (
			value reflect.Value
			desc  string
		)
		switch node := el.Visitable().(type) {
		case *Value:
			value = node.Target
			desc = node.ShortSummary()

		case *Flag:
			value = node.Target
			desc = node.ShortSummary()

		case *Application:
			value = node.Target
			desc = node.Name

		case *Node:
			value = node.Target
			desc = node.Path()
		}
		if validate := isValidatable(value); validate != nil {
			if err := validate.Validate(); err != nil {
				return Errors().DecodeValueError(desc, err)
			}
		}
	}
	for _, resolver := range c.combineResolvers() {
		if err := resolver.Validate(c.Model); err != nil {
			return err
		}
	}
	for _, path := range c.Path {
		var value *Value
		switch {
		case path.Flag != nil:
			value = path.Flag.Value

		case path.Positional != nil:
			value = path.Positional
		}
		if value != nil && value.Tag.Enum != "" {
			if err := checkEnum(value, value.Target); err != nil {
				return err
			}
		}
		if err := checkMissingFlags(path.Flags); err != nil {
			return err
		}
	}
	// Check the terminal node.
	node := c.Selected()
	if node == nil {
		node = c.Model.Node
	}

	// Find the deepest positional argument, so we can check if all required positionals have been provided.
	positionals := 0
	for _, path := range c.Path {
		if path.Positional != nil {
			positionals = path.Positional.Position + 1
		}
	}

	if err := checkMissingChildren(node); err != nil {
		return err
	}
	if err := checkMissingPositional(positionals, node.Positional); err != nil {
		return err
	}
	if err := checkXorDuplicates(c.Path); err != nil {
		return err
	}

	if node.Type == ArgumentNode {
		value := node.Argument
		if value.Required && !value.Set {
			return Errors().IsRequired(node.Summary())
		}
	}
	return nil
}

// Flags returns the accumulated available flags.
func (c *Context) Flags() (flags []*Flag) {
	for _, trace := range c.Path {
		flags = append(flags, trace.Flags...)
	}
	return
}

// Command returns the full command path.
func (c *Context) Command() string {
	var (
		trace   *Path
		command []string
	)

	for _, trace = range c.Path {
		switch {
		case trace.Positional != nil:
			command = append(command, "<"+trace.Positional.Name+">")

		case trace.Argument != nil:
			command = append(command, "<"+trace.Argument.Name+">")

		case trace.Command != nil:
			command = append(command, trace.Command.Name)
		}
	}

	return strings.Join(command, delimiterSpace)
}

// AddResolver adds a context-specific resolver.
//
// This is most useful in the BeforeResolve() hook.
func (c *Context) AddResolver(resolver Resolver) {
	c.resolvers = append(c.resolvers, resolver)
}

// FlagValue returns the set value of a flag if it was encountered and exists, or its default value.
func (c *Context) FlagValue(flag *Flag) any {
	for _, trace := range c.Path {
		if trace.Flag == flag {
			v, ok := c.values[trace.Flag.Value]
			if !ok {
				break
			}
			return v.Interface()
		}
	}
	if flag.Target.IsValid() {
		return flag.Target.Interface()
	}
	return flag.DefaultValue.Interface()
}

// Reset recursively resets values to defaults (as specified in the grammar) or the zero value.
func (c *Context) Reset() error {
	return Visit(c.Model.Node, func(node Visitable, next Next) error {
		if value, ok := node.(*Value); ok {
			return next(value.Reset())
		}
		return next(nil)
	})
}

func (c *Context) endParsing() {
	var args []string

	for {
		token := c.scan.Pop()
		if token.Type == EOLToken {
			break
		}
		args = append(args, token.String())
	}
	// Note: tokens must be pushed in reverse order.
	for i := range args {
		c.scan.PushTyped(args[len(args)-1-i], PositionalArgumentToken)
	}
}

func (c *Context) trace(node *Node) (err error) {
	var (
		flagNode   *Node
		positional int
		flags      []*Flag
	)

	flagNode = node
	if node.DefaultCmd != nil && node.DefaultCmd.Tag.Default == cmdWithArgs {
		// Add flags of the default command if the current node has one
		// and that default command allows args / flags without explicitly
		// naming the command on the CLI.
		flagNode = node.DefaultCmd
	}
	for _, group := range flagNode.AllFlags(false) {
		flags = append(flags, group...)
	}

	if node.Passthrough {
		c.endParsing()
	}

	for !c.scan.Peek().IsEOL() {
		token := c.scan.Peek()
		switch token.Type {
		case UntypedToken:
			switch v := token.Value.(type) {
			case string:

				switch {
				case v == delimiterDash:
					fallthrough
				default: // nolint
					c.scan.Pop()
					c.scan.PushTyped(token.Value, PositionalArgumentToken)

				// Indicates end of parsing. All remaining arguments are treated as positional arguments only.
				case v == delimiterDoubleDash:
					c.scan.Pop()
					c.endParsing()

				// Long flag.
				case strings.HasPrefix(v, delimiterDoubleDash):
					c.scan.Pop()
					// Parse it and push the tokens.
					parts := strings.SplitN(v[2:], "=", 2)
					if len(parts) > 1 {
						c.scan.PushTyped(parts[1], FlagValueToken)
					}
					c.scan.PushTyped(parts[0], FlagToken)

				// Short flag.
				case strings.HasPrefix(v, delimiterDash):
					c.scan.Pop()
					// Note: tokens must be pushed in reverse order.
					if tail := v[2:]; tail != "" {
						c.scan.PushTyped(tail, ShortFlagTailToken)
					}
					c.scan.PushTyped(v[1:2], ShortFlagToken)
				}
			default:
				c.scan.Pop()
				c.scan.PushTyped(token.Value, PositionalArgumentToken)
			}

		case ShortFlagTailToken:
			c.scan.Pop()
			// Note: tokens must be pushed in reverse order.
			if tail := token.String()[1:]; tail != "" {
				c.scan.PushTyped(tail, ShortFlagTailToken)
			}
			c.scan.PushTyped(token.String()[0:1], ShortFlagToken)

		case FlagToken:
			if err = c.parseFlag(flags, token.String()); err != nil {
				return
			}

		case ShortFlagToken:
			if err = c.parseFlag(flags, token.String()); err != nil {
				return
			}

		case FlagValueToken:
			return Errors().UnexpectedFlagArgument(token.Value)

		case PositionalArgumentToken:
			var candidates []string

			// Ensure we've consumed all positional arguments.
			if positional < len(node.Positional) {
				arg := node.Positional[positional]

				if arg.Passthrough {
					c.endParsing()
				}

				if err = arg.Parse(c.scan, c.getValue(arg)); err != nil {
					return err
				}
				c.Path = append(c.Path, &Path{
					Parent:     node,
					Positional: arg,
				})
				positional++
				break
			}

			// Assign token value to a branch name if tagged as an alias
			// will be ignored in the case of an existing command
			cmds := make(map[string]bool)
			for _, branch := range node.Children {
				if branch.Type == CommandNode {
					cmds[branch.Name] = true
				}
			}
			for _, branch := range node.Children {
				for _, a := range branch.Aliases {
					_, ok := cmds[a]
					if token.Value == a && !ok {
						token.Value = branch.Name
						break
					}
				}
			}

			// After positional arguments have been consumed, check commands next...
			for _, branch := range node.Children {
				if branch.Type == CommandNode && !branch.Hidden {
					candidates = append(candidates, branch.Name)
				}
				if branch.Type == CommandNode && branch.Name == token.Value {
					c.scan.Pop()
					c.Path = append(c.Path, &Path{
						Parent:  node,
						Command: branch,
						Flags:   branch.Flags,
					})
					return c.trace(branch)
				}
			}

			// Finally, check arguments.
			for _, branch := range node.Children {
				if branch.Type == ArgumentNode {
					arg := branch.Argument
					if err := arg.Parse(c.scan, c.getValue(arg)); err == nil {
						c.Path = append(c.Path, &Path{
							Parent:   node,
							Argument: branch,
							Flags:    branch.Flags,
						})
						return c.trace(branch)
					}
				}
			}

			// If there is a default command that allows args and nothing else
			// matches, take the branch of the default command
			if node.DefaultCmd != nil && node.DefaultCmd.Tag.Default == cmdWithArgs {
				c.Path = append(c.Path, &Path{
					Parent:  node,
					Command: node.DefaultCmd,
					Flags:   node.DefaultCmd.Flags,
				})
				return c.trace(node.DefaultCmd)
			}

			return findPotentialCandidates(token.String(), candidates, Errors().UnexpectedArgument(token.String()))
		default:
			return Errors().UnexpectedToken(token)
		}
	}
	return c.maybeSelectDefault(flags, node)
}

// End of the line, check for a default command, but only if we're not displaying help,
// otherwise we'd only ever display the help for the default command.
func (c *Context) maybeSelectDefault(flags []*Flag, node *Node) error {
	for _, flag := range flags {
		if flag.Name == helpName && flag.Set {
			return nil
		}
	}
	if node.DefaultCmd != nil {
		c.Path = append(c.Path, &Path{
			Parent:  node.DefaultCmd,
			Command: node.DefaultCmd,
			Flags:   node.DefaultCmd.Flags,
		})
	}
	return nil
}

// Resolve walks through the traced path, applying resolvers to any unset flags.
func (c *Context) Resolve() error {
	resolvers := c.combineResolvers()
	if len(resolvers) == 0 {
		return nil
	}

	var inserted []*Path
	for _, path := range c.Path {
		for _, flag := range path.Flags {
			// Flag has already been set on the command-line.
			if _, ok := c.values[flag.Value]; ok {
				continue
			}

			// Pick the last resolved value.
			var selected any
			for _, resolver := range resolvers {
				s, err := resolver.Resolve(c, path, flag)
				if err != nil {
					return Errors().DecodeValueError(flag.ShortSummary(), err)
				}
				if s == nil {
					continue
				}
				selected = s
			}

			if selected == nil {
				continue
			}

			scan := Scan().PushTyped(selected, FlagValueToken)
			delete(c.values, flag.Value)
			err := flag.Parse(scan, c.getValue(flag.Value))
			if err != nil {
				return err
			}
			inserted = append(inserted, &Path{
				Flag:     flag,
				Resolved: true,
			})
		}
	}
	c.Path = append(c.Path, inserted...)
	return nil
}

// Combine application-level resolvers and context resolvers.
func (c *Context) combineResolvers() []Resolver {
	var resolvers = []Resolver{}

	resolvers = append(resolvers, c.Kong.resolvers...)
	resolvers = append(resolvers, c.resolvers...)

	return resolvers
}

func (c *Context) getValue(value *Value) reflect.Value {
	v, ok := c.values[value]
	if !ok {
		v = reflect.New(value.Target.Type()).Elem()
		switch v.Kind() {
		case reflect.Ptr:
			v.Set(reflect.New(v.Type().Elem()))
		case reflect.Slice:
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		case reflect.Map:
			v.Set(reflect.MakeMap(v.Type()))
		default:
		}
		c.values[value] = v
	}
	return v
}

// ApplyDefaults if they are not already set.
func (c *Context) ApplyDefaults() error {
	return Visit(c.Model.Node, func(node Visitable, next Next) error {
		var value *Value
		switch node := node.(type) {
		case *Flag:
			value = node.Value
		case *Node:
			value = node.Argument
		case *Value:
			value = node
		default:
		}
		if value != nil {
			if err := value.ApplyDefault(); err != nil {
				return err
			}
		}
		return next(nil)
	})
}

// Apply traced context to the target grammar.
func (c *Context) Apply() (string, error) {
	var path []string

	for _, trace := range c.Path {
		var value *Value
		switch {
		case trace.App != nil:
		case trace.Argument != nil:
			path = append(path, "<"+trace.Argument.Name+">")
			value = trace.Argument.Argument
		case trace.Command != nil:
			path = append(path, trace.Command.Name)
		case trace.Flag != nil:
			value = trace.Flag.Value
		case trace.Positional != nil:
			path = append(path, "<"+trace.Positional.Name+">")
			value = trace.Positional
		default:
			panic("unsupported path ?!")
		}
		if value != nil {
			value.Apply(c.getValue(value))
		}
	}

	return strings.Join(path, delimiterSpace), nil
}

func (c *Context) parseFlag(flags []*Flag, match string) (err error) {
	candidates := []string{}
	for _, flag := range flags {
		long := delimiterDoubleDash + flag.Name
		short := delimiterDash + string(flag.Short)
		neg := "--no-" + flag.Name
		candidates = append(candidates, long)
		if flag.Short != 0 {
			candidates = append(candidates, short)
		}
		if short != match && long != match && !(match == neg && flag.Tag.Negatable) {
			continue
		}
		// Found a matching flag.
		c.scan.Pop()
		if match == neg && flag.Tag.Negatable {
			flag.Negated = true
		}
		err := flag.Parse(c.scan, c.getValue(flag.Value))
		if err != nil {
			var expected *expectedError
			if errors.As(err, &expected) && expected.token.InferredType().IsAny(FlagToken, ShortFlagToken) {
				return fmt.Errorf("%s; perhaps try %s=%q?", err.Error(), flag.ShortSummary(), expected.token)
			}
			return err
		}
		if flag.Negated {
			value := c.getValue(flag.Value)
			value.SetBool(!value.Bool())
			flag.Value.Apply(value)
		}
		c.Path = append(c.Path, &Path{Flag: flag})
		return nil
	}
	return findPotentialCandidates(match, candidates, Errors().UnknownFlag(match))
}

// RunNode calls the Run() method on an arbitrary node.
//
// This is useful in conjunction with Visit(), for dynamically running commands.
//
// Any passed values will be bindable to arguments of the target Run() method. Additionally,
// all parent nodes in the command structure will be bound.
func (c *Context) RunNode(node *Node, binds ...any) (err error) {
	type targetMethod struct {
		node   *Node
		method reflect.Value
		binds  bindings
	}
	methodBinds := c.Kong.bindings.clone().add(binds...).add(c).merge(c.bindings)
	methods := []targetMethod{}
	for i := 0; node != nil; i, node = i+1, node.Parent {
		method := getMethod(node.Target, "Run")
		methodBinds = methodBinds.clone()
		for p := node; p != nil; p = p.Parent {
			methodBinds = methodBinds.add(p.Target.Addr().Interface())
		}
		if method.IsValid() {
			methods = append(methods, targetMethod{node, method, methodBinds})
		}
	}
	if len(methods) == 0 {
		return fmt.Errorf("no Run() method found in hierarchy of %s", c.Selected().Summary())
	}
	_, err = c.Apply()
	if err != nil {
		return err
	}

	for _, method := range methods {
		if err = callMethod("Run", method.node.Target, method.method, method.binds); err != nil {
			return err
		}
	}
	return nil
}

// Run executes the Run() method on the selected command, which must exist.
//
// Any passed values will be bindable to arguments of the target Run() method. Additionally,
// all parent nodes in the command structure will be bound.
func (c *Context) Run(binds ...any) (err error) {
	node := c.Selected()
	if node == nil {
		if len(c.Path) > 0 {
			selected := c.Path[0].Node()
			if selected.Type == ApplicationNode {
				method := getMethod(selected.Target, "Run")
				if method.IsValid() {
					return c.RunNode(selected, binds...)
				}
			}
		}
		return Errors().NoCommandSelected()
	}
	return c.RunNode(node, binds...)
}

// PrintUsage to Kong's stdout.
//
// If summary is true, a summarised version of the help will be output.
func (c *Context) PrintUsage(summary bool) error {
	options := c.helpOptions
	options.Summary = summary
	return c.help(options, c)
}

func checkMissingFlags(flags []*Flag) (err error) {
	const delimiterOr = ` or `
	var (
		flag         *Flag
		missing, fgs []string
		xor          string
		xorGroupSet  = map[string]bool{}
		xorGroup     = map[string][]string{}
	)

	for _, flag = range flags {
		if flag.Set {
			for _, xor = range flag.Xor {
				xorGroupSet[xor] = true
			}
		}
		if !flag.Required || flag.Set {
			continue
		}
		if len(flag.Xor) > 0 {
			for _, xor = range flag.Xor {
				if xorGroupSet[xor] {
					continue
				}
				xorGroup[xor] = append(xorGroup[xor], flag.Summary())
			}
		} else {
			missing = append(missing, flag.Summary())
		}
	}
	for xor, fgs = range xorGroup {
		if !xorGroupSet[xor] && len(fgs) > 1 {
			missing = append(missing, strings.Join(fgs, delimiterOr))
		}
	}
	if len(missing) == 0 {
		return
	}
	sort.Strings(missing)
	err = Errors().MissingFlags(strings.Join(missing, delimiterCommaSpace))

	return
}

func checkMissingChildren(node *Node) (err error) {
	var (
		arg         *Positional
		child       *Node
		missing     []string
		missingArgs []string
	)

	for _, arg = range node.Positional {
		if arg.Required && !arg.Set {
			missingArgs = append(missingArgs, arg.Summary())
		}
	}
	if len(missingArgs) > 0 {
		missing = append(missing, strconv.Quote(strings.Join(missingArgs, delimiterSpace)))
	}
	for _, child = range node.Children {
		if child.Hidden || (child.Argument != nil && !child.Argument.Required) {
			continue
		}
		if child.Argument != nil {
			missing = append(missing, strconv.Quote(child.Summary()))
		} else {
			missing = append(missing, strconv.Quote(child.Name))
		}
	}
	err = checkErrorOfMissingChildren(missing)

	return
}

func checkErrorOfMissingChildren(missing []string) (err error) {
	const maxMissing = 5

	if len(missing) > maxMissing {
		missing = append(missing[:maxMissing], onyOther)
	}
	switch len(missing) {
	case 0:
		return
	case 1:
		err = Errors().Expected(missing[0])
	default:
		err = Errors().ExpectedOneOf(strings.Join(uniqueQuotedNotEmpty(missing), delimiterCommaSpace))
	}

	return
}

// If we're missing any positionals and they're required, return an error.
func checkMissingPositional(positional int, values []*Value) (err error) {
	var (
		missing []string
		arg     *Value
		ok      bool
	)

	// All the positionals are in.
	if positional >= len(values) {
		return
	}
	// We're low on supplied positionals, but the missing one is optional.
	if !values[positional].Required {
		return
	}
	for ; positional < len(values); positional++ {
		arg = values[positional]
		if arg.Tag.Env != "" {
			if _, ok = os.LookupEnv(arg.Tag.Env); ok {
				continue
			}
		}
		missing = append(missing, "<"+arg.Name+">")
	}
	if len(missing) == 0 {
		return
	}
	err = Errors().MissingPositionalArguments(strings.Join(missing, delimiterSpace))

	return
}

func checkEnum(value *Value, target reflect.Value) (err error) {
	var (
		enums, enumSlice []string
		enum             string
		tmp              string
		i                int
	)

	switch target.Kind() {
	case reflect.Slice, reflect.Array:
		for i = 0; i < target.Len(); i++ {
			if err = checkEnum(value, target.Index(i)); err != nil {
				break
			}
		}
	case reflect.Map, reflect.Struct:
		err = Errors().EnumSliceOrValue()
	default:
		enumSlice = value.EnumSlice()
		tmp = fmt.Sprintf("%v", target)
		for _, enum = range enumSlice {
			if enum == tmp {
				return
			}
			enums = append(enums, fmt.Sprintf("%q", enum))
		}
		err = Errors().MustBeOneOfButGot(value.ShortSummary(), strings.Join(enums, delimiterComma), target.Interface())
	}

	return
}

func checkPassthroughArg(target reflect.Value) bool {
	typ := target.Type()
	switch typ.Kind() {
	case reflect.Slice:
		return typ.Elem().Kind() == reflect.String
	default:
		return false
	}
}

func checkXorDuplicates(paths []*Path) error {
	for _, path := range paths {
		seen := map[string]*Flag{}
		for _, flag := range path.Flags {
			if !flag.Set {
				continue
			}
			for _, xor := range flag.Xor {
				if seen[xor] != nil {
					return fmt.Errorf("--%s and --%s can't be used together", seen[xor].Name, flag.Name)
				}
				seen[xor] = flag
			}
		}
	}
	return nil
}

func findPotentialCandidates(needle string, haystack []string, inErr error) (err error) {
	var (
		closestCandidates []string
		candidate         string
	)

	if err = inErr; len(haystack) == 0 {
		return
	}
	closestCandidates = make([]string, 0, len(haystack))
	for _, candidate = range haystack {
		if strings.HasPrefix(candidate, needle) || levenshtein(candidate, needle) <= 2 {
			closestCandidates = append(closestCandidates, fmt.Sprintf("%q", candidate))
		}
	}
	switch len(closestCandidates) {
	case 0:
		return
	case 1:
		err = Errors().HelperErrorDidYouMean(inErr, closestCandidates[0])
	default:
		err = Errors().HelperErrorDidYouMeanOneOf(inErr, closestCandidates)
	}

	return
}

type validatable interface{ Validate() error }

func isValidatable(v reflect.Value) validatable {
	if !v.IsValid() || (v.Kind() == reflect.Ptr || v.Kind() == reflect.Slice || v.Kind() == reflect.Map) && v.IsNil() {
		return nil
	}
	if validate, ok := v.Interface().(validatable); ok {
		return validate
	}
	if v.CanAddr() {
		return isValidatable(v.Addr())
	}
	return nil
}
