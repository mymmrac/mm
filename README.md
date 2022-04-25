# :diamond_shape_with_a_dot_inside: mm

Simple CLI math expression evaluator.

Order agnostic expression evaluator (order of operations determined only by order of operators and not
operands itself) meaning that `++ 1` and `1 ++` essentially the same expression. 
It also supports constants like `Pi` and other.

**mm** uses repl to interact with user with live results and error highlighting.

## :jigsaw: Get Started

Install using `go install`:

```shell
go install github.com/mymmrac/mm@latest
```

> Note: Make sure to add `$GOPATH/bin` into `$PATH`

Start repl and type some expressions:

```shell
mm

> 1 + 1

> 2 ^ 8 / 3.1

> 1 / (2.5 !ceil) -
```

## :keyboard: Shortcuts

- `Enter` - evaluate expression
- `Up`, `Tab` - previews executed expression
- `Down`, `Shift+Tab` - next executed expression
- `Shift+Tab` - use result of last expression as input (only if input empty)
- `Esc` - exit if input is empty, or clean input
- `Crtl+c` - force quit

## :zap: Operators

- `+` Addition
- `-` Subtraction
- `*` Multiplication
- `/` Division
- `^` Power (only integer powers)
- `@` Nth Root (only integer roots)
- `%` Mod (only integers)
- `-` Unary Minus
- `++` Increment
- `--` Decrement
- `!abs` Abs
- `!round` Round
- `!floor` Floor
- `!ceil` Ceil

## :book: Constants

- `Pi` - 3.1415926...
- `e` - 2.7182818...

## :closed_lock_with_key: License

**mm** is distributed under [MIT licence](LICENSE).
