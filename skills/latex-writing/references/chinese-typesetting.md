# LaTeX Chinese Typesetting

## Recommended Approach

For Chinese documents, prefer XeLaTeX or LuaLaTeX with a `ctex` document class.

```latex
\documentclass[UTF8,a4paper,12pt]{ctexart}
\begin{document}
中文、中英文混排和数学公式 \(E = mc^2\) 可以直接写。
\end{document}
```

Common `ctex` document classes:

- `ctexart`: articles and short reports.
- `ctexrep`: longer reports.
- `ctexbook`: books.
- `ctexbeamer`: Chinese slides.

## Use `ctex` with a Standard Class

If an existing school, journal, or project template already defines the document class, load `ctex` instead:

```latex
\documentclass{article}
\usepackage{ctex}
```

## Font Configuration

With XeLaTeX or LuaLaTeX, CJK fonts can be configured directly:

```latex
\setCJKmainfont{Noto Serif CJK SC}
\setCJKsansfont{Noto Sans CJK SC}
\setCJKmonofont{Noto Sans Mono CJK SC}
```

Available fonts differ across machines. If compilation fails because a font is missing, either switch to an installed font or remove explicit font settings and use the `ctex` defaults.

## Mixed Chinese-English Text

`ctex` supports mixed Chinese, English, and math content:

```latex
这是中文。This is English. 数学公式为 \(F = ma\)。
```

## Chinese Abstract and Table of Contents

`ctexart` localizes names such as Abstract and Table of Contents into Chinese.

```latex
\begin{abstract}
这里是摘要。
\end{abstract}

\tableofcontents
```

## Not Recommended

Do not prefer pdfLaTeX for Chinese-first documents. pdfLaTeX can be used with `CJKutf8`, but every Chinese segment needs a `CJK` environment, so it is better for English documents with small Chinese snippets.

```latex
\documentclass{article}
\usepackage{CJKutf8}
\begin{document}
\begin{CJK*}{UTF8}{gbsn}
中文片段。
\end{CJK*}
\end{document}
```

## Chinese Technical Report Example

```latex
\documentclass[UTF8,a4paper,12pt]{ctexart}
\usepackage{geometry}
\usepackage{amsmath}
\usepackage{graphicx}
\usepackage{booktabs}
\usepackage{tabularx}
\usepackage{hyperref}

\geometry{margin=2.5cm}

\title{技术报告标题}
\author{作者}
\date{\today}

\begin{document}
\maketitle
\begin{abstract}
这里写摘要。
\end{abstract}

\tableofcontents

\section{背景}
这里写正文。

\section{设计}
这里写设计内容。

\section{结论}
这里写结论。

\end{document}
```

## Sources

- https://www.overleaf.com/learn/latex/Chinese
- https://ctan.org/pkg/ctex
