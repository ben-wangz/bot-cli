---
name: latex-writing
description: LaTeX writing, Chinese technical reports, formulas, figures, tables, bibliography, command-line compilation, and troubleshooting. Use when creating, editing, compiling, or debugging .tex documents.
metadata:
  skill-version: "1.0.0"
---

# LaTeX Writing

Use this skill for LaTeX document authoring and maintenance, especially Chinese technical reports, research papers, project documents, formulas, figures, tables, citations, and command-line compilation.

## Quick Task Routing

- Document structure, sections, lists, references, packages, and special characters: `references/syntax-cheatsheet.md`.
- Mathematical notation, equations, alignment, matrices, and symbols: `references/math.md`.
- Figures, tables, floating environments, captions, labels, and long tables: `references/figures-and-tables.md`.
- Bibliography setup, `.bib` entries, citation commands, and compile order: `references/bibliography.md`.
- Chinese documents, `ctex`, XeLaTeX/LuaLaTeX, fonts, and mixed Chinese-English text: `references/chinese-typesetting.md`.
- LaTeX environment installation on Ubuntu 24.04 and Fedora 42: `references/environment-installation.md`.
- Terminal compilation, `latexmk`, XeLaTeX, Biber, and cleanup: `references/command-line-compile.md`.
- Common LaTeX failures and fixes: `references/troubleshooting.md`.

## Operating Principles

- Prefer XeLaTeX or LuaLaTeX for Chinese documents.
- Prefer `ctexart`, `ctexrep`, `ctexbook`, or `ctexbeamer` for Chinese-first documents.
- Prefer `latexmk -xelatex main.tex` for routine command-line compilation.
- Put `hyperref` after most packages unless a template says otherwise.
- Use labels consistently: `sec:`, `fig:`, `tab:`, and `eq:`.
- For figures and tables, place `\label{...}` after `\caption{...}`.
- Do not use `$$ ... $$`; use `\[ ... \]` or equation environments.
- Keep source files UTF-8 encoded.

## Output Checklist

- The document compiles with the intended engine.
- Chinese text renders correctly when present.
- Cross references, figure references, table references, equation references, and citations resolve.
- Figures use stable relative paths and filenames without spaces.
- Wide tables are handled with `tabularx`, `p{...}`, or another suitable layout.
- Auxiliary files are not treated as source artifacts unless explicitly needed.
