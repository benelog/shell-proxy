# Writing for Shell Proxy

These rules cover the prose of this repository: the Markdown files, the release notes, and anything written to GitHub such as an issue comment, a pull request description or a release body.

## ATX headings only

Write every heading with leading `#` marks.
Never underline a heading with `===` or `---`.

    no    Writing for Shell Proxy
          =======================

    yes   # Writing for Shell Proxy

## No em dashes

Never use an em dash (`—`), and never use an en dash (`–`) in its place.
Where one would have gone, use the punctuation the sentence actually calls for, or join the clauses with a conjunction.

    no    The server keeps nothing on disk — no configuration file, no session store.
    yes   The server keeps nothing on disk: no configuration file, no session store.

    no    In stateless mode this makes no difference — one command in, one result out — but the exit code still matters.
    yes   In stateless mode this makes no difference, since one command goes in and one result comes out, but the exit code still matters.

A colon introduces what follows, a comma or a semicolon separates clauses, and parentheses hold an aside.
One of those fits every place an em dash would have.

## One sentence, one line

In Markdown, write each sentence on a line of its own and let it run as long as it needs to.
Do not wrap a paragraph to a column width: a sentence is never broken across two lines, and two sentences never share one.
The rendered page is unaffected, and a diff then shows the sentence that changed rather than every line the rewrap moved.

    no    Interactive mode opens a real PTY and streams it to the browser: it
          runs a full login shell, carries the terminal size across, and ends
          when the shell exits. Stateless mode does none of that.

    yes   Interactive mode opens a real PTY and streams it to the browser: it runs a full login shell, carries the terminal size across, and ends when the shell exits.
          Stateless mode does none of that.

Inside a list item, continuation sentences line up with the text of the item:

    - `--interactive` turns the PTY mode on.
      Without it the server stays stateless.

Code blocks, tables and headings are left exactly as they are; the rule is about prose.
