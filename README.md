## base32k
[![GoDoc](https://godoc.org/github.com/grandchild/base32k?status.svg)](https://godoc.org/github.com/grandchild/base32k)
[![License](https://img.shields.io/github/license/grandchild/base32k.svg)](https://creativecommons.org/publicdomain/zero/1.0/)

*base32k* is a slightly whimsical binary-to-text encoding, which transforms
raw binary data into (possibly obscene combinations of) UTF-8 characters from
the CJK and Hangul unicode blocks. Its alphabet consists of 2^15 = 32768
characters, hence the name. In comparison to other encodings like *base64* or
*base122*, this does not save space in terms of bytes, but it is smaller than
those two in terms of characters. This is only useful when for some reason the
medium (* cough * twitter * cough * ) is character-limited rather than byte-
limited.

#### Example
    $ echo testing | base32k
    整棦릥茻l

    $ echo 整棦릥茻l | base32k -d
    testing

#### Installation
##### Library
    go get github.com/grandchild/base32k

##### Executable Binary
    go get github.com/grandchild/base32k/base32k

#### Encoding Ratio
*base32k* has an encoding ratio of 15 bits per unicode glyph, which amounts to
a ratio of 15/24 (0.625) plus one byte padding in 14 out of 15 cases. This
makes it a worse encoding than *base64*, which has a 3/4 (0.75) and *base122*,
which has 7/8 (0.875).

A twitter message may be 280 characters long, but only 140 CJK glyphs. Still,
this encoding slightly outperforms base64 and even base122 in the space
available in a single tweet:

<table>
<thead>
    <tr>
        <td></td>
        <td>space ratio</td><td>char ratio</td><td>bytes per tweet</td>
    </tr>
</thead>
<tbody>
    <tr><td>base64  </td><td> 0.75  </td><td>  6 </td><td> 210 </td></tr>
    <tr><td>base122 </td><td> 0.875 </td><td>  7 </td><td> 245 </td></tr>
    <tr><td>base32k </td><td> 0.625 </td><td> 15 </td><td> 256 </td></tr>
    <tr><td></td><td colspan=3 align="center">
        ( more is better for all columns )
    </td></tr>
</tbody>
</table>

*base32k* outperforming *base122* on twitter results from the fact that
twitter counts a CJK or Hangul glyph as two characters, whereas in UTF-8 it's
actually 3 characters. This gives us, in effect, an encoding ratio of 15/16
over *base122*'s 7/8, a slight advantage.

So, given good-enough font coverage of the basic multilingual unicode plane,
this can be used to transmit data in situations where characters are limited,
rather than disk space.

#### Stability
This implementation will run out of memory when en-/decoding very large chunks
of data (several gigabytes). But since this is aimed at character-limited
settings this is not likely to be an issue.
