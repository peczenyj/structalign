#!/usr/bin/env python3
"""Render structalign's ANSI-colored output to a terminal-style PNG.

Handles only the SGR codes structalign emits: reset(0), bold(1), dim(2),
red(31), green(32), cyan(36). Tabs expand to 4 spaces.
"""
import re
import sys
from PIL import Image, ImageDraw, ImageFont

SRC, OUT = sys.argv[1], sys.argv[2]

# --- theme (Tomorrow-Night-ish on a dark background) ---
BG       = (29, 31, 33)       # #1d1f21
BAR      = (40, 43, 48)       # title bar
DEFAULT  = (197, 200, 198)    # #c5c8c6
RED      = (224, 108, 117)    # deletions
GREEN    = (152, 195, 121)    # additions
CYAN     = (86, 182, 194)     # header
DOTS     = [(255, 95, 86), (255, 189, 46), (39, 201, 63)]
PROMPT   = "$ structalign ./_example"

FONT_DIR = "/usr/share/fonts/truetype/dejavu"
SIZE     = 30
reg  = ImageFont.truetype(f"{FONT_DIR}/DejaVuSansMono.ttf", SIZE)
bold = ImageFont.truetype(f"{FONT_DIR}/DejaVuSansMono-Bold.ttf", SIZE)

CHAR_W = reg.getlength("M")
asc, desc = reg.getmetrics()
LINE_H = asc + desc + 6
PAD = 28
BAR_H = 56

def dim(c):
    return tuple(int(v * 0.55) for v in c)

def parse_line(line):
    """-> list of (text, color, is_bold)"""
    color, b, d = DEFAULT, False, False
    out = []
    for tok in re.split(r'(\x1b\[[0-9;]*m)', line):
        if not tok:
            continue
        m = re.fullmatch(r'\x1b\[([0-9;]*)m', tok)
        if m:
            for p in (m.group(1) or "0").split(';'):
                p = p or "0"
                if p == "0":   color, b, d = DEFAULT, False, False
                elif p == "1": b = True
                elif p == "2": d = True
                elif p == "22": b = d = False
                elif p == "31": color = RED
                elif p == "32": color = GREEN
                elif p == "36": color = CYAN
                elif p in ("39", "0"): color = DEFAULT
            continue
        c = dim(color) if d else color
        out.append((tok.replace('\t', '    '), c, b))
    return out

raw_lines = open(SRC, encoding="utf-8").read().rstrip("\n").split("\n")
# Synthetic prompt line on top for context.
lines = [[(PROMPT, dim(DEFAULT), False)]] + [[]] + [parse_line(l) for l in raw_lines]

def line_cols(segs):
    return sum(len(t) for t, _, _ in segs)

cols = max((line_cols(s) for s in lines), default=20)
W = int(cols * CHAR_W) + 2 * PAD
H = BAR_H + len(lines) * LINE_H + 2 * PAD

img = Image.new("RGB", (W, H), BG)
d = ImageDraw.Draw(img)

# title bar + traffic lights
d.rectangle([0, 0, W, BAR_H], fill=BAR)
for i, col in enumerate(DOTS):
    cx = PAD + i * 34
    cy = BAR_H // 2
    d.ellipse([cx - 9, cy - 9, cx + 9, cy + 9], fill=col)

y = BAR_H + PAD
for segs in lines:
    x = PAD
    for text, color, is_bold in segs:
        f = bold if is_bold else reg
        d.text((x, y), text, font=f, fill=color)
        x += len(text) * CHAR_W
    y += LINE_H

img.save(OUT)
print(f"wrote {OUT}  ({W}x{H})")
