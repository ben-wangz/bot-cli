# LaTeX Environment Installation

This reference covers practical package installation for writing and compiling LaTeX documents, especially Chinese technical reports that use XeLaTeX, `ctex`, `graphicx`, `booktabs`, `tabularx`, `hyperref`, and `biblatex` with Biber.

## Ubuntu 24.04

Install the recommended packages:

```bash
sudo apt update
sudo apt install -y \
  latexmk \
  texlive-xetex \
  texlive-lang-chinese \
  texlive-latex-recommended \
  texlive-latex-extra \
  texlive-science \
  biber \
  fonts-noto-cjk
```

Package purpose:

- `latexmk`: automatic multi-pass compilation.
- `texlive-xetex`: XeLaTeX engine and core XeTeX support.
- `texlive-lang-chinese`: Chinese language support, including `ctex`-related files.
- `texlive-latex-recommended`: common LaTeX packages.
- `texlive-latex-extra`: extra packages often used by reports and templates.
- `texlive-science`: science and math-related packages.
- `biber`: bibliography backend for `biblatex`.
- `fonts-noto-cjk`: broadly available CJK fonts.

Verify the installation:

```bash
xelatex --version
latexmk -v
biber --version
```

Compile a minimal Chinese document:

```bash
cat > main.tex <<'EOF'
\documentclass[UTF8,a4paper,12pt]{ctexart}
\begin{document}
Chinese LaTeX compilation test.
\end{document}
EOF
latexmk -xelatex main.tex
```

If `main.pdf` is generated and Chinese text renders correctly, the environment is usable.

## Fedora 42

Install the recommended packages:

```bash
sudo dnf install -y \
  latexmk \
  texlive-xetex \
  texlive-ctex \
  texlive-amsmath \
  texlive-graphics \
  texlive-booktabs \
  texlive-tabularx \
  texlive-longtable \
  texlive-geometry \
  texlive-hyperref \
  texlive-enumitem \
  texlive-biber \
  texlive-fandol
```

If `texlive-tabularx` or `texlive-longtable` is not available as a separate package in the configured Fedora repositories, install the broader collection that provides it:

```bash
sudo dnf install -y texlive-collection-latexextra
```

Package purpose:

- `latexmk`: automatic multi-pass compilation.
- `texlive-xetex`: XeLaTeX engine and core XeTeX support.
- `texlive-ctex`: Chinese document classes and package support.
- `texlive-amsmath`: mathematical typesetting.
- `texlive-graphics`: `graphicx` and graphics support.
- `texlive-booktabs`: professional table rules.
- `texlive-tabularx`: fixed-width adaptive tables.
- `texlive-longtable`: multi-page tables.
- `texlive-geometry`: page margin control.
- `texlive-hyperref`: hyperlinks and PDF metadata.
- `texlive-enumitem`: list spacing and list formatting control.
- `texlive-biber`: bibliography backend for `biblatex`.
- `texlive-fandol`: default Chinese fonts commonly used by `ctex`.

Verify the installation:

```bash
xelatex --version
latexmk -v
biber --version
```

Compile a minimal Chinese document:

```bash
cat > main.tex <<'EOF'
\documentclass[UTF8,a4paper,12pt]{ctexart}
\begin{document}
Chinese LaTeX compilation test.
\end{document}
EOF
latexmk -xelatex main.tex
```

If `main.pdf` is generated and Chinese text renders correctly, the environment is usable.

## Minimal Containers

In minimal containers, `sudo` may be unavailable because commands already run as root. In that case, remove `sudo` from the installation commands.

If only a small subset is needed, install at least the engine, `ctex`, graphics support, and `latexmk`:

```bash
# Ubuntu 24.04
apt update
apt install -y latexmk texlive-xetex texlive-lang-chinese texlive-latex-recommended fonts-noto-cjk

# Fedora 42
dnf install -y latexmk texlive-xetex texlive-ctex texlive-graphics texlive-fandol
```

## Troubleshooting Installation

- `xelatex: command not found`: install `texlive-xetex`.
- `latexmk: command not found`: install `latexmk`.
- `File ctexart.cls not found`: install `texlive-lang-chinese` on Ubuntu or `texlive-ctex` on Fedora.
- `biber: command not found`: install `biber` on Ubuntu or `texlive-biber` on Fedora.
- Chinese glyphs are missing: install CJK fonts, such as `fonts-noto-cjk` on Ubuntu or `texlive-fandol` on Fedora, or configure an installed CJK font explicitly.
