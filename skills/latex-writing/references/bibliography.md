# LaTeX Bibliography

## Recommended Setup

For new documents, prefer `biblatex` with `biber`:

```latex
\usepackage[
  backend=biber,
  style=numeric,
  sorting=none
]{biblatex}
\addbibresource{references.bib}
```

Citation in body text:

```latex
Einstein's paper is cited as Reference~\cite{einstein1905}.
```

Print the bibliography:

```latex
\printbibliography
```

## Complete Example

```latex
\documentclass{article}
\usepackage[
  backend=biber,
  style=numeric,
  sorting=none
]{biblatex}
\addbibresource{references.bib}

\begin{document}
This is a citation~\cite{einstein1905}.

\printbibliography
\end{document}
```

## `.bib` Entries

`references.bib`:

```bibtex
@article{einstein1905,
  author  = {Albert Einstein},
  title   = {Zur Elektrodynamik bewegter K{\"o}rper},
  journal = {Annalen der Physik},
  volume  = {322},
  number  = {10},
  pages   = {891--921},
  year    = {1905},
  doi     = {10.1002/andp.19053221004}
}

@book{dirac1981,
  author    = {Paul Adrien Maurice Dirac},
  title     = {The Principles of Quantum Mechanics},
  publisher = {Clarendon Press},
  year      = {1981}
}

@online{ctan,
  author  = {{CTAN}},
  title   = {The Comprehensive TeX Archive Network},
  url     = {https://ctan.org/},
  urldate = {2026-06-23}
}
```

## Common Citation Commands

```latex
\cite{key}
\cite{key1,key2}
\parencite{key}   % Supported by some styles.
\textcite{key}    % Supported by some styles.
```

## Common Styles

```latex
style=numeric
style=alphabetic
style=authoryear
```

Sorting:

```latex
sorting=none  % Citation order.
sorting=ynt   % Year, name, title.
sorting=nty   % Name, title, year.
```

## Add Bibliography to the Table of Contents

```latex
\printbibliography[heading=bibintoc,title={References}]
```

## Compile Order

For `biblatex` with `biber`, the typical manual order is:

```text
xelatex main.tex
biber main
xelatex main.tex
xelatex main.tex
```

Many editors and `latexmk` can automate this process.

## Source

- https://www.overleaf.com/learn/latex/Bibliography_management_in_LaTeX
