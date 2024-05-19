# :diamond_shape_with_a_dot_inside: mm

Simple CLI math expression evaluator.

**mm** uses repl to interact with user with live results and error highlighting, but immediate mode is also supported.

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

> 1 / ceil(2.5 + 4 / (abs(sin(5))))
```

## :keyboard: Shortcuts

- `Enter` - evaluate expression
- `Up`, `Tab` - previews executed expression
- `Down`, `Shift+Tab` - next executed expression
- `Shift+Tab` - use the result of last expression as input (only if input empty)
- `Esc` - exit if input is empty, or clean input
- `Crtl+c` - force quit

## :zap: Operators

### Binary

- `+` Addition
- `-` Subtraction
- `*` Multiplication
- `/` Division
- `//` Floor division
- `^` Power
- `%` Modulo

### Unary

- `+` Plus
- `-` Minus

## :hash: Functions

- `sqrt/1` Square root
- `abs/1` Absolute value
- `round/1` Round to integer
- `round/2` Round with precision
- `roundUp/1` Round up to integer
- `roundUp/2` Round up with precision
- `floor/1` Floor
- `ceil/1` Ceil
- `sin/1` Sine
- `cos/1` Cosine
- `tan/1` Tangent
- `atan/1` Arc tangent
- `rad/1` To radians
- `min/2` Minimum
- `max/2` Maximum
- `rand/0` Random value [0, 1)

> Note: `<name>/N` means that `<name>` is called with `N` arguments

## :book: Constants

- `Pi` - 3.1415926...
- `e` - 2.7182818...

## :closed_lock_with_key: License

**mm** is distributed under [MIT licence](LICENSE).
