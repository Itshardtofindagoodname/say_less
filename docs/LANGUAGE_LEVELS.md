# Say Less Language Levels

This document explains the three levels involved in Say Less: **Neography**,
the **interpreter**, and the **compiler**. The levels describe different
concerns; they are not three names for the same program.

## At A Glance

| Level | Main question | Say Less representation | Current status |
|-------|---------------|-------------------------|----------------|
| Neography | How does a human write the language? | UTF-8 `.sl` source, keywords, expressions, indentation, and syntax | Implemented as the language surface |
| Interpreter | How can the source run now? | Tokens -> AST -> direct evaluation | Current execution model |
| Compiler | How can source become an executable artifact? | AST -> typed IR -> bytecode or native code | Planned Phase 2 |

The relationship is:

```text
Neography:     source.sl
                    |
                    v
Lexer:         tokens
                    |
                    v
Parser:        AST
                 /    \\
                /      \\
Interpreter     Compiler (planned)
directly       AST -> IR -> bytecode/native
evaluates AST        |
                    v
                VM/native execution
```

## 1. Neography: The Written Language

Neography is the visible notation of Say Less: the symbols, words, layout, and
conventions a programmer uses to express a program. It is a language-design
level, not a runtime engine.

In Say Less, Neography includes:

- `.sl` files encoded as UTF-8.
- Keywords such as `fn`, `if`, `while`, `return`, `use`, `struct`, and `test`.
- Literals such as integers, floats, strings, booleans, and `none`.
- Operators such as `+`, `==`, `and`, `or`, and assignment operators.
- Significant newlines and four-space indentation for blocks.
- Concise function calls, where parentheses are optional in supported forms.
- Comments beginning with `#`.

Example:

```sl
fn add(a, b)
    return a + b

result = add 2 3
print result
```

Neography does not execute this program. It only gives the programmer a stable
way to write it. The lexer is the first implementation component that reads
this notation and turns it into tokens.

## 2. Interpreter: Running the Meaning Directly

An interpreter executes a program by examining its parsed representation during
runtime. Say Less currently uses a tree-walking interpreter.

The current path is:

```text
.sl source
    -> lexer.Tokenize()
    -> parser.Parse()
    -> parser.Program (AST)
    -> eval.Interpreter.Run()
    -> result or runtime error
```

The evaluator walks AST nodes and performs their actions. For example, it:

- Evaluates literals and expressions into runtime `Value` objects.
- Stores variables in nested `Environment` values.
- Creates functions that retain their defining environment.
- Executes control-flow nodes such as `if`, `while`, and `for`.
- Calls Say Less functions and Go-backed builtins.

This is why `sale run file.sl` works without producing a separate executable:
the CLI reads the source, lexes it, parses it, and passes the AST directly to
`eval.Interpreter`.

### Strengths of the Current Interpreter

- Simple implementation and fast iteration while the language evolves.
- Useful runtime errors can be tied directly to source positions.
- No bytecode format or VM is required yet.
- Builtins can call Go code directly.

### Tradeoffs

- The AST is revisited during execution, adding runtime overhead.
- The program generally needs the Say Less runtime and Go implementation to run.
- Optimization is limited compared with a compiler pipeline.

## 3. Compiler: Translating Before Execution

A compiler transforms a Say Less program into another executable representation
before the program runs. That representation may be bytecode for a Say Less VM,
native machine code, or another supported target.

The planned compiler path is:

```text
.sl source
    -> lexer
    -> parser
    -> semantic analysis
    -> typed AST
    -> SSA-based IR
    -> optimization
    -> bytecode or native code
    -> VM or native execution
```

The compiler can perform work ahead of time, including type checks, constant
folding, dead-code removal, and layout decisions. A bytecode compiler would
produce instructions for a Say Less virtual machine. A native compiler would
produce an executable for a target operating system and CPU.

The compiler is **not currently implemented** in this repository. The
architecture document describes it as a Phase 2 project. The present evaluator
must therefore be treated as the source of truth for runtime behavior.

## 4. How the Levels Fit Together

These levels answer different questions:

| Question | Responsible level |
|----------|-------------------|
| What does `fn add(a, b)` look like? | Neography |
| Is the source structurally valid? | Lexer and parser |
| What value does `add 2 3` produce right now? | Interpreter |
| Can the program be checked or optimized before running? | Compiler pipeline |
| Where does a compiled instruction execute? | VM or native runtime |

The lexer and parser are shared front-end stages. Both the interpreter and a
future compiler can consume the AST produced by those stages. This keeps the
language notation independent from the execution strategy.

## 5. Current Repository Mapping

| Responsibility | Location |
|----------------|----------|
| Command-line entry point | `cmd/sale/main.go` |
| Tokenization | `internal/lexer/lexer.go` |
| Parsing | `internal/parser/parser.go` |
| AST definitions | `internal/parser/ast.go` |
| Tree-walking evaluation | `internal/eval/eval.go` and `internal/eval/expressions.go` |
| Go-backed builtins | `internal/eval/builtins.go` |
| Planned compiler design | Described in `docs/ARCHITECTURE.md`; implementation pending |

## Terminology Rule

When documenting or discussing Say Less, use these terms precisely:

- **Neography** means the written syntax and notation of the language.
- **Interpreter** means the current AST evaluator that runs Say Less directly.
- **Compiler** means the planned translator from AST to bytecode or native code.
- **Runtime** means the support needed while a program executes, including
  values, environments, builtins, and eventually a VM.

In short: Neography is how Say Less is written, the interpreter is how the
current implementation runs it, and the compiler is a future way to translate
it for execution.