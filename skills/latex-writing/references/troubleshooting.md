# LaTeX Troubleshooting

## Chinese Text Is Missing or Garbled

Check first:

- Use XeLaTeX or LuaLaTeX.
- Use `ctexart`, `ctexrep`, `ctexbook`, or `\usepackage{ctex}`.
- Keep source files UTF-8 encoded.

Minimal verification:

```latex
\documentclass[UTF8]{ctexart}
\begin{document}
Chinese text test.
\end{document}
```

## `Undefined control sequence`

Common causes:

- A command is misspelled.
- A required package is missing, such as `graphicx` for `\includegraphics`.
- A command is not supported by the current template.

Example fix:

```latex
\usepackage{graphicx}
```

## `Missing $ inserted`

Common cause: math commands are used in text mode, or `_` and `^` are used outside math mode.

Wrong:

```latex
x_1
```

Correct:

```latex
\(x_1\)
```

For a literal underscore:

```latex
file\_name
```

## Cross References Show `??`

Fixes:

- Confirm that `\label{...}` and `\ref{...}` match exactly.
- Compile at least twice.
- For figures and tables, put `\label{...}` after `\caption{...}`.

```latex
\caption{Example figure}
\label{fig:example}
```

## Image Not Found

Check:

- The path is relative to the main `.tex` file.
- The filename has no spaces.
- `graphicx` is loaded.
- The compiler supports the image format.

```latex
\usepackage{graphicx}
\includegraphics[width=\textwidth]{figures/example.png}
```

## Table Is Too Wide

Options:

- Use the `X` column in `tabularx` for automatic wrapping.
- Use `p{...}` for fixed-width columns.
- Reduce font size only when layout changes are insufficient.
- Use `pdflscape` for landscape pages.

```latex
\usepackage{tabularx}

\begin{tabularx}{\textwidth}{lXr}
\toprule
Name & Description & Status \\
\midrule
Task & A long description that needs wrapping & Done \\
\bottomrule
\end{tabularx}
```

## Bibliography Is Missing

Check:

- `\addbibresource{references.bib}` exists.
- Body text contains `\cite{key}`.
- `biber` has been run when using `biblatex` with `backend=biber`.
- The `.bib` key matches the key used by `\cite{...}`.

Minimal setup:

```latex
\usepackage[backend=biber,style=numeric]{biblatex}
\addbibresource{references.bib}
```

## Special Characters Fail Compilation

Escape these characters in text mode:

```latex
\# \$ \% \& \_ \{ \}
```

Backslash itself:

```latex
\textbackslash{}
```

## `hyperref` Issues

Usually load `hyperref` after most packages:

```latex
\usepackage{hyperref}
```

If Chinese bookmarks are garbled, prefer XeLaTeX with `ctex` and avoid pdfLaTeX for Chinese-first documents.

## Sources

- https://www.overleaf.com/learn/latex/Learn_LaTeX_in_30_minutes
- https://www.overleaf.com/learn/latex/Chinese
- https://www.overleaf.com/learn/latex/Tables
- https://www.overleaf.com/learn/latex/Bibliography_management_in_LaTeX
