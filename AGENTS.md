# Repository rules

## Project understanding

- When you need to understand the project, start by reading [INDEX.md](INDEX.md). It links to all relevant documentation. Browse only the files relevant to the task at hand.

- [Taskfile.yml](Taskfile.yml) is used to run most common commands. Check it first before running a command from scratch.

## Software Engineering principles

- **Skateboard before Lamborghini**: If an idea or concept needs to be validated, create first an MVP instead of a full fledged feature, validate the assumptions (e.g. if it can work, if it's useful, etc...) and only then expand it.
- **ABT**: Always Be Testing. Write extremely high-quality tests, use them extensively and follow Test Driven Design.
- **Locality of Behavior**: It's easier to understand parts of code that work together if they are colocated. Try to keep this in mind when writing.
- **YAGNI**: You Aren't Gonna Need It. Use coding best practices but code for the now. There's always time to add stuff later.
- **Modularity**: Build for flexibility and optionality, so it's easy to add, remove or change features.
- **Two-way door decisions**: Make decisions that can be easily reverted (like a two-way door, you can go in and go out).

## OpenSpec rules

- When it is required of the user to continue an OpenSpec loop via a command (e.g. /opsx-continue), use the "question" tool to prompt the user with the command to use next, including an open-ended option where the user can type it in.
