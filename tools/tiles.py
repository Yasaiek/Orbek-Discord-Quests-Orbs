#!/usr/bin/env python3
"""Colourful icon tiles: a rounded square with a gradient and a white sign, smooth at any size.

    python3 tools/tiles.py ico  <out dir> <app>   tray menu icons (.ico, 16–32 px)   app: orbek | loorek
    python3 tools/tiles.py png  <out dir> <app>   Discord activity pictures (.png, 1024 px)
"""
import math, os, sys
from PIL import Image, ImageDraw, ImageFilter

S = 1024  # drawn here, scaled down

def grad_tile(c1, c2, radius=0.26, shine=True):
    g = Image.new('RGBA', (S, S))
    px = g.load()
    a, b = [int(c1[i:i + 2], 16) for i in (1, 3, 5)], [int(c2[i:i + 2], 16) for i in (1, 3, 5)]
    for y in range(S):
        for x in range(0, S, 4):
            t = (x + y) / (2 * S)
            col = tuple(int(a[i] + (b[i] - a[i]) * t) for i in range(3)) + (255,)
            for k in range(4): px[x + k, y] = col
    mask = Image.new('L', (S, S), 0)
    ImageDraw.Draw(mask).rounded_rectangle((0, 0, S - 1, S - 1), int(S * radius), fill=255)
    out = Image.new('RGBA', (S, S), (0, 0, 0, 0)); out.paste(g, (0, 0), mask)
    if shine:  # soft light from the top
        hl = Image.new('RGBA', (S, S), (0, 0, 0, 0)); d = ImageDraw.Draw(hl)
        d.ellipse((-S * .3, -S * .9, S * 1.3, S * .45), fill=(255, 255, 255, 38))
        hl = hl.filter(ImageFilter.GaussianBlur(S * .09))
        m2 = Image.new('L', (S, S), 0); m2.paste(mask)
        out = Image.alpha_composite(out, Image.composite(hl, Image.new('RGBA', (S, S), (0, 0, 0, 0)), m2))
    return out

W = 70  # line width of the signs (at 1024)
WH = (255, 255, 255, 255)

def sign(name, d):
    c, w = WH, W
    if name == 'open':
        d.rounded_rectangle((220, 270, 804, 754), 70, outline=c, width=w); d.line((220, 400, 804, 400), fill=c, width=w)
        for x in (310, 400): d.ellipse((x - 28, 307, x + 28, 363), fill=c)
    elif name == 'mini':
        for x, y in ((240, 240), (548, 240), (240, 548), (548, 548)): d.rounded_rectangle((x, y, x + 236, y + 236), 60, fill=c)
    elif name == 'stop':
        d.rounded_rectangle((300, 300, 724, 724), 80, fill=c)
    elif name == 'update':
        d.arc((240, 240, 784, 784), 210, 500, fill=c, width=w)
        d.polygon([(760, 200), (790, 440), (560, 380)], fill=c)
    elif name == 'site':
        d.ellipse((220, 220, 804, 804), outline=c, width=w); d.ellipse((390, 220, 634, 804), outline=c, width=w - 10)
        d.line((220, 512, 804, 512), fill=c, width=w - 10)
    elif name == 'folder':
        d.rounded_rectangle((210, 330, 814, 760), 60, fill=c); d.rounded_rectangle((210, 270, 470, 400), 50, fill=c)
    elif name == 'quit':
        d.arc((250, 270, 774, 794), -55, 235, fill=c, width=w + 6); d.line((512, 190, 512, 500), fill=c, width=w + 14)
    elif name == 'usage':
        d.rounded_rectangle((300, 300, 724, 724), 70, outline=c, width=w); d.rounded_rectangle((430, 430, 594, 594), 30, fill=c)
        for p in (400, 512, 624):
            for a, b in (((p, 190), (p, 300)), ((p, 724), (p, 834)), ((190, p), (300, p)), ((724, p), (834, p))): d.line(a + b, fill=c, width=w - 16)
    elif name == 'video':
        d.rounded_rectangle((190, 270, 834, 754), 80, outline=c, width=w); d.polygon([(450, 390), (450, 634), (660, 512)], fill=c)
    elif name == 'play':
        d.rounded_rectangle((170, 330, 854, 714), 190, outline=c, width=w)
        d.line((290, 522, 450, 522), fill=c, width=w - 6); d.line((370, 442, 370, 602), fill=c, width=w - 6)
        d.ellipse((600, 440, 670, 510), fill=c); d.ellipse((680, 530, 750, 600), fill=c)
    elif name == 'activity':
        d.line([(180, 540), (360, 540), (450, 300), (580, 760), (670, 520), (844, 520)], fill=c, width=w, joint='curve')
        for p in ((180, 540), (844, 520)): d.ellipse((p[0] - 35, p[1] - 35, p[0] + 35, p[1] + 35), fill=c)
    elif name == 'stream':
        d.rounded_rectangle((190, 250, 834, 680), 70, outline=c, width=w); d.line((390, 800, 634, 800), fill=c, width=w); d.line((512, 690, 512, 800), fill=c, width=w)
        d.polygon([(450, 360), (450, 570), (620, 465)], fill=c)
    elif name == 'points':
        d.ellipse((230, 230, 794, 794), outline=c, width=w); d.polygon([(512, 330), (560, 460), (700, 470), (590, 555), (630, 690), (512, 610), (394, 690), (434, 555), (324, 470), (464, 460)], fill=c)
    elif name == 'drops':
        d.polygon([(512, 180), (760, 520), (512, 860), (264, 520)], outline=c, width=w)
        d.line((512, 300, 512, 740), fill=c, width=w - 20)
    elif name == 'kick':
        d.line((360, 260, 360, 764), fill=c, width=w + 20); d.line((380, 520, 680, 260), fill=c, width=w + 20); d.line((440, 470, 690, 764), fill=c, width=w + 20)

PAL = {  # tile colours per sign (from, to)
    'open': ('#6d7dff', '#9b5cff'), 'mini': ('#22c1c3', '#4f7dff'), 'stop': ('#ff6b6b', '#d83a6e'),
    'update': ('#2ed39a', '#18a0c9'), 'site': ('#3aa8ff', '#5865f2'), 'folder': ('#ffb547', '#ff7a3d'),
    'quit': ('#ff5e7a', '#c2255c'), 'usage': ('#ffcc4d', '#ff8a3d'), 'video': ('#ff8a5b', '#ff4f8b'),
    'play': ('#5b8cff', '#8a5bff'), 'activity': ('#29d3b4', '#3a8bff'), 'stream': ('#b05bff', '#ff4fb0'),
    'points': ('#9b5cff', '#5865f2'), 'drops': ('#36b4f2', '#3a6bff'), 'kick': ('#53fc18', '#1fb85a'),
}

def tile(name):
    im = grad_tile(*PAL[name]); sg = Image.new('RGBA', (S, S), (0, 0, 0, 0)); sign(name, ImageDraw.Draw(sg))
    sh = Image.new('RGBA', (S, S), (0, 0, 0, 0)); sh.paste((0, 0, 0, 70), (0, 14), sg.split()[3]); sh = sh.filter(ImageFilter.GaussianBlur(14))
    return Image.alpha_composite(Image.alpha_composite(im, sh), sg)

def dot(color):
    im = Image.new('RGBA', (S, S), (0, 0, 0, 0)); d = ImageDraw.Draw(im)
    d.ellipse((300, 300, 724, 724), fill=color); return im

if __name__ == '__main__':
    mode, out, app = sys.argv[1], sys.argv[2], sys.argv[3]
    os.makedirs(out, exist_ok=True)
    if mode == 'ico':
        names = ['open', 'mini', 'stop', 'update', 'site', 'folder', 'quit', 'usage'] + (['points', 'drops', 'kick'] if app == 'loorek' else [])
        for n in names:
            tile(n).save(os.path.join(out, n + '.ico'), sizes=[(16, 16), (20, 20), (24, 24), (32, 32)])
        dot((62, 207, 142, 255)).save(os.path.join(out, 'on.ico'), sizes=[(16, 16), (20, 20), (24, 24), (32, 32)])
        dot((139, 143, 163, 255)).save(os.path.join(out, 'off.ico'), sizes=[(16, 16), (20, 20), (24, 24), (32, 32)])
    else:
        for n in ['video', 'play', 'activity', 'stream']:
            tile(n).resize((1024, 1024), Image.LANCZOS).save(os.path.join(out, n + '.png'))
    print(mode, 'icons in', out)
