# LaTeX Figures and Tables

## Figures

Load the package in the preamble:

```latex
\usepackage{graphicx}
```

Insert a figure:

```latex
\begin{figure}[htbp]
  \centering
  \includegraphics[width=0.75\textwidth]{figures/example}
  \caption{Example figure}
  \label{fig:example}
\end{figure}
```

Reference it:

```latex
As shown in Figure~\ref{fig:example}.
```

Recommendations:

- Avoid spaces and repeated dots in image filenames.
- The file extension can often be omitted.
- Common widths include `width=0.75\textwidth` and `width=\linewidth`.
- Put `\label{...}` after `\caption{...}`.

## Floating Positions

Common placement options:

| Option | Meaning |
| --- | --- |
| `h` | Try here |
| `t` | Top of page |
| `b` | Bottom of page |
| `p` | Float-only page |
| `!` | Relax LaTeX internal constraints |

Common combination:

```latex
\begin{figure}[htbp]
```

## Basic Tables

```latex
\begin{table}[htbp]
  \centering
  \caption{Example table}
  \label{tab:example}
  \begin{tabular}{lcr}
    \hline
    Name & Count & Description \\
    \hline
    A & 10 & Left aligned \\
    B & 20 & Right aligned \\
    \hline
  \end{tabular}
\end{table}
```

Column formats:

- `l`: left aligned.
- `c`: centered.
- `r`: right aligned.
- `p{3cm}`: fixed-width paragraph column.
- `|`: vertical rule.

## Professional Tables

For research documents, prefer `booktabs`:

```latex
\usepackage{booktabs}
```

```latex
\begin{table}[htbp]
  \centering
  \caption{Booktabs example}
  \label{tab:booktabs}
  \begin{tabular}{lll}
    \toprule
    Dataset & Band & Purpose \\
    \midrule
    Gaia DR3 & Optical & Astrometry \\
    DESI & Spectrum & Spectral analysis \\
    \bottomrule
  \end{tabular}
\end{table}
```

## Fixed-Width Tables

```latex
\usepackage{tabularx}
```

```latex
\begin{tabularx}{\textwidth}{lXr}
\toprule
Name & Description & Status \\
\midrule
Workflow & This is a longer description that wraps automatically & Done \\
\bottomrule
\end{tabularx}
```

## Merged Cells

Merge columns:

```latex
\multicolumn{3}{c}{Merged across three columns}
```

Merge rows with `multirow`:

```latex
\usepackage{multirow}
```

```latex
\multirow{2}{*}{Merged across two rows}
```

## Multi-Page Tables

```latex
\usepackage{longtable}
```

```latex
\begin{longtable}{ll}
\caption{Long table}\label{tab:long}\\
\toprule
Column 1 & Column 2 \\
\midrule
\endfirsthead
\toprule
Column 1 & Column 2 \\
\midrule
\endhead
Content & Content \\
\bottomrule
\end{longtable}
```

## Sources

- https://www.overleaf.com/learn/latex/Tables
- https://ctan.org/pkg/graphicx
