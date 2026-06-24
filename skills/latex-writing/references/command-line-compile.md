# LaTeX Command-Line Compilation

This reference covers manual LaTeX compilation from the terminal, without relying on an editor plugin.

## Basic Concept

LaTeX source files must be compiled before they can be read as final output:

```text
main.tex -> compile -> main.pdf
```

If the source has not changed, an existing `main.pdf` can be opened directly. Recompile only when `.tex` files, images, bibliography files, or style files change.

## Recommended Command for Chinese Documents

For Chinese documents, prefer XeLaTeX:

```bash
xelatex main.tex
```

If the document contains a table of contents, cross references, figure references, table references, or equation numbers, it usually needs at least two compilation passes:

```bash
xelatex main.tex
xelatex main.tex
```

## Recommended: `latexmk`

`latexmk` automatically determines how many passes are needed and is the preferred daily command:

```bash
latexmk -xelatex main.tex
```

Typical outputs include:

- `main.pdf`: final PDF.
- `main.aux`, `main.log`, `main.toc`, and other auxiliary files.

## Compilation with Bibliography

For `biblatex` with `biber`:

```bash
xelatex main.tex
biber main
xelatex main.tex
xelatex main.tex
```

With `latexmk`, this is usually automatic:

```bash
latexmk -xelatex main.tex
```

If bibliography output does not refresh, explicitly use Biber:

```bash
latexmk -xelatex -use-biber main.tex
```

## Cleanup

Remove intermediate files while keeping the PDF:

```bash
latexmk -c main.tex
```

Remove intermediate files and the PDF:

```bash
latexmk -C main.tex
```

If `latexmk` is unavailable, delete common auxiliary files manually:

```bash
rm -f main.aux main.log main.out main.toc main.bbl main.bcf main.blg main.run.xml
```

Be careful not to delete `.tex`, `.bib`, image files, or PDFs that need to be retained.

## Compiler Selection

| Command | Best For |
| --- | --- |
| `xelatex main.tex` | Chinese documents, mixed Chinese-English text, system fonts |
| `lualatex main.tex` | Chinese documents, modern fonts, advanced typesetting |
| `pdflatex main.tex` | English documents and traditional templates |
| `latexmk -xelatex main.tex` | Recommended daily command with automatic multi-pass compilation |

Do not prefer `pdflatex` for Chinese-first documents.

## Minimal Verification File

Create `main.tex`:

```latex
\documentclass[UTF8,a4paper,12pt]{ctexart}
\begin{document}
Chinese LaTeX compilation test.
\end{document}
```

Compile:

```bash
latexmk -xelatex main.tex
```

Or:

```bash
xelatex main.tex
```

If `main.pdf` is generated, the local LaTeX compilation environment works.

## Common Issues

- Missing `xelatex`: TeX Live, MacTeX, or MiKTeX is not installed, or the command is not in `PATH`.
- Chinese text is garbled or missing: use XeLaTeX/LuaLaTeX and `ctexart` or `\usepackage{ctex}`.
- References show `??`: compile again, or use `latexmk -xelatex main.tex`.
- Bibliography is missing: run `biber main`, or use `latexmk -xelatex -use-biber main.tex`.
- Image not found: verify that the image path is relative to the main `.tex` file and that the filename has no spaces.
