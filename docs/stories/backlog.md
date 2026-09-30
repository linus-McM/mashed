# Backlog

Moved from the untracked `todo.md` (repo health remediation, R29). Items keep
the owner's wording; lines marked `>` in the original are kept as "(flagged)".

- (flagged) Right-click menu for the file explorer: add, delete, rename files.
- (flagged) A BMAD file-reader module so a node can point to a file as an input.
- An array process that can be linked to loop and regex pattern nodes (BMAD).
- A text input box that can call a Claude process or pass text to a BMAD agent.
- A response modal every time BMAD has a question for the user:
  - A snackbar/toast that can appear from any session, from any BMAD process running in any directory. It stays in the top-right corner until the user responds, and other BMAD workflows can add to it.
  - Snackbars use the colour the user set on the home page panel and are ordered by creation time.
  - Clicking a snackbar takes the user to that repository's BMAD workspace and opens the response modal. After a text or menu response and confirmation, the response is sent back to the terminal session so it can continue.
- An options-review modal with logical flow and/or set operations that flag the user with a snackbar linking to the process flow on the canvas.
- A code review panel with summarisation that can switch between static advice and extreme-programming mode.
- Open question: how are BMAD workspace flows linked to terminal sessions?
- Get toolbar tips working in the markdown viewer.
