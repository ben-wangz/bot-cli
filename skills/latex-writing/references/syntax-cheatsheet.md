# LaTeX Syntax Cheatsheet

## Minimal Document

```latex
\documentclass{article}
\begin{document}
Hello LaTeX.
\end{document}
```

For Chinese documents, prefer:

```latex
\documentclass[UTF8,a4paper,12pt]{ctexart}
\begin{document}
Chinese text can be written directly when using `ctex`.
\end{document}
```

## Preamble

Everything before `\begin{document}` is the preamble. Use it for the document class, packages, title metadata, and page layout.

```latex
\documentclass[12pt,a4paper]{article}
\usepackage{graphicx}
\usepackage{amsmath}
\title{My Document}
\author{Author Name}
\date{\today}
```

## Title and Table of Contents

```latex
\title{Document Title}
\author{Author Name}
\date{\today}

\begin{document}
\maketitle
\tableofcontents
\end{document}
```

## Section Levels

Common sectioning commands:

```latex
\part{Part}
\chapter{Chapter}        % Supported by report/book/ctexrep/ctexbook.
\section{Section}
\subsection{Subsection}
\subsubsection{Subsubsection}
\paragraph{Paragraph Title}
```

Use the starred form for unnumbered headings:

```latex
\section*{Unnumbered Heading}
```

## Paragraphs and Line Breaks

- Leave one blank line to start a new paragraph.
- `\\` forces a line break; do not use it to simulate paragraph spacing.
- `\newline` also forces a line break.

```latex
First paragraph.

Second paragraph.
```

## Comments

```latex
% This line does not appear in the PDF.
```

## Text Styles

```latex
\textbf{Bold text}
\textit{Italic text}
\underline{Underlined text}
\emph{Emphasized text}
\texttt{Monospace text}
```

## Lists

Unordered list:

```latex
\begin{itemize}
  \item First item
  \item Second item
\end{itemize}
```

Ordered list:

```latex
\begin{enumerate}
  \item First item
  \item Second item
\end{enumerate}
```

## Cross References

The usual pattern is to place `\label{...}` at the target and refer to it with `\ref{...}` or `\pageref{...}`.

```latex
\section{System Design}\label{sec:design}

As described in Section~\ref{sec:design}.
```

Common label prefixes:

- `sec:` for sections.
- `fig:` for figures.
- `tab:` for tables.
- `eq:` for equations.

## Hyperlinks

```latex
\usepackage{hyperref}
```

Load `hyperref` after most other packages unless a template instructs otherwise.

## Common Packages

```latex
\usepackage{amsmath}   % Math.
\usepackage{graphicx}  % Figures.
\usepackage{booktabs}  % Professional tables.
\usepackage{tabularx}  % Fixed-width adaptive tables.
\usepackage{longtable} % Multi-page tables.
\usepackage{hyperref}  % Hyperlinks.
\usepackage{geometry}  % Page margins.
```

## Special Characters

These characters have special meanings and usually need escaping in text mode:

| Character | LaTeX |
| --- | --- |
| `#` | `\#` |
| `$` | `\$` |
| `%` | `\%` |
| `&` | `\&` |
| `_` | `\_` |
| `{` | `\{` |
| `}` | `\}` |
| `\` | `\textbackslash{}` |

## Source

- https://www.overleaf.com/learn/latex/Learn_LaTeX_in_30_minutes
