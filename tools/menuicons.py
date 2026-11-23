#!/usr/bin/env python3
"""Draws the small icons of the tray menu (ICO, 16–32 px): python3 tools/menuicons.py <out dir> <#accent>"""
import math, os, sys
from PIL import Image, ImageDraw

out, accent = sys.argv[1], sys.argv[2]
S = 256  # drawn large, scaled down smooth
W = 18   # line width at 256 px

def rgb(h):
    h = h.lstrip('#'); return tuple(int(h[i:i + 2], 16) for i in (0, 2, 4)) + (255,)

def icon(name, draw, color=accent):
    im = Image.new('RGBA', (S, S), (0, 0, 0, 0)); d = ImageDraw.Draw(im); draw(d, rgb(color))
    im.save(os.path.join(out, name + '.ico'), sizes=[(16, 16), (20, 20), (24, 24), (32, 32)])

def arc(d, box, a0, a1, c):
    d.arc(box, a0, a1, fill=c, width=W)

def window(d, c):
    d.rounded_rectangle((28, 48, 228, 208), 26, outline=c, width=W)
    d.line((28, 96, 228, 96), fill=c, width=W)
    for x in (60, 92): d.ellipse((x - 9, 63, x + 9, 81), fill=c)

def mini(d, c):
    for x, y in ((40, 40), (140, 40), (40, 140), (140, 140)): d.rounded_rectangle((x, y, x + 76, y + 76), 18, outline=c, width=W)

def stop(d, c):
    d.rounded_rectangle((64, 64, 192, 192), 26, fill=c)

def update(d, c):
    arc(d, (40, 40, 216, 216), 200, 500 - 40, c)
    d.polygon([(206, 40), (214, 112), (146, 92)], fill=c)

def globe(d, c):
    d.ellipse((36, 36, 220, 220), outline=c, width=W)
    d.ellipse((92, 36, 164, 220), outline=c, width=W)
    d.line((36, 128, 220, 128), fill=c, width=W)

def power(d, c):
    arc(d, (44, 52, 212, 220), -50, 230, c)
    d.line((128, 28, 128, 120), fill=c, width=W + 4)

def chip(d, c):
    d.rounded_rectangle((64, 64, 192, 192), 22, outline=c, width=W)
    d.rectangle((104, 104, 152, 152), fill=c)
    for p in (96, 128, 160):
        d.line((p, 28, p, 64), fill=c, width=W - 4); d.line((p, 192, p, 228), fill=c, width=W - 4)
        d.line((28, p, 64, p), fill=c, width=W - 4); d.line((192, p, 228, p), fill=c, width=W - 4)

def dot(d, c):
    d.ellipse((72, 72, 184, 184), fill=c)

def play(d, c):
    d.polygon([(84, 52), (204, 128), (84, 204)], fill=c)

def folder(d, c):
    d.rounded_rectangle((32, 72, 224, 204), 20, outline=c, width=W)
    d.line((32, 96, 112, 96), fill=c, width=W); d.line((112, 96, 132, 72), fill=c, width=W)

os.makedirs(out, exist_ok=True)
icon('open', window); icon('mini', mini); icon('stop', stop, '#f0616d'); icon('update', update); icon('site', globe)
icon('quit', power, '#f0616d'); icon('usage', chip); icon('on', dot, '#3ecf8e'); icon('off', dot, '#8b8fa3')
icon('play', play); icon('folder', folder)
print('icons in', out)
