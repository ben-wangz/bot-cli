# LaTeX Writing Reference Overview

This reference set contains practical LaTeX guidance for technical reports, research papers, and project documentation. It is adapted from common Overleaf and CTAN guidance and is optimized for command-line workflows and Chinese technical documents.

## Recommended Compile Flow

For Chinese documents, prefer XeLaTeX or LuaLaTeX. The default command-line choice is:

```bash
latexmk -xelatex main.tex
```

If `latexmk` is unavailable, compile directly:

```bash
xelatex main.tex
xelatex main.tex
```

## Minimal Chinese Template

```latex
\documentclass[UTF8,a4paper,12pt]{ctexart}
\usepackage{amsmath}
\usepackage{graphicx}
\usepackage{booktabs}
\usepackage{hyperref}

\title{Document Title}
\author{Author Name}
\date{\today}

\begin{document}
\maketitle
\tableofcontents

\section{Introduction}
This is the document body.

\end{document}
```

## Reference Map

- Syntax and document structure: `syntax-cheatsheet.md`.
- Math: `math.md`.
- Figures and tables: `figures-and-tables.md`.
- Bibliography: `bibliography.md`.
- Chinese typesetting: `chinese-typesetting.md`.
- Command-line compilation: `command-line-compile.md`.
- Troubleshooting: `troubleshooting.md`.

## Authoritative Sources

- Overleaf Learn LaTeX in 30 minutes: https://www.overleaf.com/learn/latex/Learn_LaTeX_in_30_minutes
- Overleaf Mathematical expressions: https://www.overleaf.com/learn/latex/Mathematical_expressions
- Overleaf Tables: https://www.overleaf.com/learn/latex/Tables
- Overleaf Bibliography management: https://www.overleaf.com/learn/latex/Bibliography_management_in_LaTeX
- Overleaf Chinese: https://www.overleaf.com/learn/latex/Chinese
- CTAN ctex: https://ctan.org/pkg/ctex
- CTAN amsmath: https://ctan.org/pkg/amsmath
- CTAN graphicx: https://ctan.org/pkg/graphicx
