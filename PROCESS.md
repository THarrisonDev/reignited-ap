# Process

## Brains > Bots

Human ingenuity and creative expression will outlast the creativity of LLMs
and generative systems. The vision behind this project is human, and no model
shares authorship of it.

That conviction is not a rejection of the tools. There is no denying the
usefulness of those same systems in fulfilling the vision of the former. Used
in service of a human design, they accelerate execution; used as a substitute
for it, they erase the thinking that makes the work worth doing.

So the rule here is simple: **the human owns the result.** Every design
decision, every line that ships, and every merge is the author's
responsibility. AI is leverage, never the author.

## Environment

Development happens in a Neovim environment (LazyVim-based) with the standard
tools of direct coding. Alongside them sit a pair of editor plugins that keep
AI assistance within reach without taking over the editor: one for the Pi
coding harness and one for OpenCode.

## Workflow

Every unit of work starts with the developer, not a prompt:

1. **Spec and design by hand.** The developer manually outlines the spec and
   the design for any and all work. This step is never delegated.
2. **Scaffold with OpenCode.** The spec is handed to OpenCode not to fully
   implement, but to scaffold the required files, functions, variables, and
   other structure. It lays out the shape; it does not decide it.
3. **Implement directly.** From that scaffold, the developer implements the
   project directly in the editor as the primary author.
4. **Pi for well-spec'd blocks.** For blocks of code whose spec is already
   tight, the developer uses the more primitive, direct tooling of Pi, driven
   by locally hosted models (Qwen 3 Coder) for the majority of that work.
5. **Larger-scale tasks on Ollama cloud.** Limited larger-scale jobs —
   reformatting, documentation, and the like — run on DeepSeek 4.1 Flash in the
   Ollama cloud environment.

The through-line: the human sets the spec and owns the result at every step.
The models fill in structure and grunt work; they never set direction.

## For contributors

Contributors are held to the same expectation. AI-assisted contributions are
welcome, but the human author owns the result: correctness, legibility, tests,
and the decision to merge. Say so in your pull request — disclosure is there
to keep review honest, not to reject a change. We will not be accepting any
code generated in a blind agentic manner and will automatically reject any
PRs directly attributed to AI tools.
