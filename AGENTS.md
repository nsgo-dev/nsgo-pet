# nsgo-pet

A desktop pet on NSGo and the embeddable pet engine behind it. Read [ARCHITECTURE.md](ARCHITECTURE.md) before coding, then the relevant subjects.

## Principles

- Keep the engine in `pet/` free of pets, assets, user-facing text and app policy; anything a host would want to decide belongs in the app or in `Options`.
- Never add a pet without its `NOTICE.md` and a license that allows shipping it; list it under Pets in both READMEs.
- A pack's clips ship in both forms, `webm/` and `mov/`, or not at all.

## Working with the user

- Do not create sub-Agents or Git worktrees; work directly in the current checkout.
- Ask before changing or removing anything the engine exports, tagging an engine version (`pet/v<version>`) or adding a dependency.
- Commit, push or open pull requests only when the user asks.
- An explicit request to publish a named version covers the release commit, the `v<version>` tag and the pushes. Requests to prepare a release, edit documents or record release rules do not.

## Verification

Run both test suites from the README's [Development](README.md#development) section after every change. Anything that touches playback, windows or the pointer also needs a manual run with `go tool nsgo dev`.

## Documentation

- Update a document only when a fact its reader needs changes.
- Avoid secondary labels and descriptions that restate a heading or value; let the layout explain.
- Write only what code and comments cannot show: purpose, boundaries and reasons, not names, fields, defaults or step-by-step behavior.
- Keep each document within its budget; check its size after editing it and trim it before handing off when it is over.
- Record completed user-visible outcomes as one-clause `Unreleased` bullets with references crediting each pull request's author, such as `(#12 by @alice)`, ordered by type and merging related ones. Before the first release, write `Unreleased` as the first version's notes.
- Repair documents you touch instead of only appending to them.
- Keep `README.md` and `README.zh-CN.md` equivalent; everything else, including code comments, is English.
- Keep documentation on the code branch and reconcile it during merges.
