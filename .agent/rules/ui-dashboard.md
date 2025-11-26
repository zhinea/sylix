---
trigger: glob
globs: ui/dashboard/**/*
---

Dashboard implementation of Sylix Database Management, build on react-router v7, tailwindcss, and shadcn.

# Follow these rules
- reuse components shadcn as many u can on "ui/dashboard/app/components"
- use `pnpm` as package manager.
- use the component concept. Each component is in its own file, with no more than 300 lines in each file.
- use TypeScript for type safety.
- use hook concept for state management and side effects.

IMPORTANT: There should be no redundant, inefficient, or duplicate code.

IMPORTANT: CREATE MAINTENABLE AND CLEAN CODE.