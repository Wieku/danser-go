# danser current osu! difficulty/PP runtime

This helper exposes the current osu!lazer difficulty and performance calculator
to danser over a local JSON-lines pipe. It is intentionally kept out of the
render loop: beatmap difficulty attributes are calculated once, while live PP
is updated from danser's simulated hit results.

Install production dependencies with Node.js 24 and pnpm:

```text
pnpm install --prod --frozen-lockfile
```

The runtime uses `@tosuapp/lazer-calculator-prebuilt`, an LGPL-3.0 binding of
osu!lazer's calculator. Its license is installed with the package under
`node_modules`.
