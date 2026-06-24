# LaTeX Math

## Recommended Package

```latex
\usepackage{amsmath}
```

`amsmath` is the core AMS-LaTeX math package and is appropriate for serious mathematical typesetting.

## Inline Math

Recommended:

```latex
The mass-energy relation is \(E = mc^2\).
```

Also valid:

```latex
$E = mc^2$
\begin{math}E = mc^2\end{math}
```

## Display Math

Unnumbered:

```latex
\[
E = mc^2
\]
```

Numbered:

```latex
\begin{equation}\label{eq:energy}
E = mc^2
\end{equation}
```

Reference:

```latex
Equation~\ref{eq:energy} gives the mass-energy relation.
```

Do not use `$$ ... $$`; it is not recommended LaTeX style.

## Superscripts and Subscripts

```latex
x^2
a_i
T^{i_1 i_2}_{j_1 j_2}
```

Wrap multiple characters in braces:

```latex
x^{10}
a_{ij}
```

## Fractions, Roots, Integrals, and Sums

```latex
\frac{a}{b}
\sqrt{x^2 + 1}
\int_0^1 f(x)\,dx
\sum_{i=1}^{n} i
\prod_{i=1}^{n} x_i
```

## Greek Letters

```latex
\alpha \beta \gamma \delta \epsilon
\theta \lambda \mu \pi \rho \sigma \omega
\Delta \Omega \Sigma \Pi
```

## Common Relations and Operators

```latex
\times \otimes \oplus
\cup \cap
\subset \supset \subseteq \supseteq
\leq \geq \neq \approx
```

Use commands for mathematical functions instead of ordinary italic letters:

```latex
\sin x
\cos x
\log x
\exp x
```

## Multi-Line Alignment

```latex
\begin{align}
a &= b + c \\
  &= d + e
\end{align}
```

Unnumbered:

```latex
\begin{align*}
a &= b + c \\
  &= d + e
\end{align*}
```

## Matrices

```latex
\[
\begin{pmatrix}
1 & 2 \\
3 & 4
\end{pmatrix}
\]
```

Common matrix environments:

- `matrix`: no brackets.
- `pmatrix`: parentheses.
- `bmatrix`: square brackets.
- `vmatrix`: single vertical bars.
- `Vmatrix`: double vertical bars.

## Sources

- https://www.overleaf.com/learn/latex/Mathematical_expressions
- https://ctan.org/pkg/amsmath
