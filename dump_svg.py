from typing import TypedDict

import cairo
import json


class ExportedNote(TypedDict):
    kind: int
    seconds: float
    track: float
    width: float


class ExportedEdge(TypedDict):
    a: int
    b: int
    connected: bool


class ExportedData(TypedDict):
    notes: list[ExportedNote]
    edges: list[ExportedEdge]


if __name__ == '__main__':
    SECONDS_SCALE = 10
    SECONDS_GAP = 1
    TRACKS_SCALE = 11
    TRACKS_GAP = 0.5
    NOTE_HEIGHT = 0.2
    NOTE_GAP = 0.04
    with open('./out.json') as raw_data:
        data: ExportedData = json.load(raw_data)
        max_seconds = max(data['notes'], key=lambda n: n['seconds'])['seconds']
        SURFACE_WIDTH = TRACKS_SCALE + TRACKS_GAP * 2
        SURFACE_HEIGHT = max_seconds * SECONDS_SCALE + SECONDS_GAP * 2
        with cairo.SVGSurface('out.svg', SURFACE_WIDTH, SURFACE_HEIGHT) as surface:
            ctx = cairo.Context(surface)
            for e in data['edges']:
                if e['connected']:
                    ctx.set_source_rgb(1, 0, 0)
                    ctx.set_line_width(0.05)
                    ctx.set_dash([])
                else:
                    ctx.set_source_rgba(0.5, 0.5, 0.5, 0.5)
                    ctx.set_line_width(0.05)
                    ctx.set_dash([0.1, 0.1])
                a = data['notes'][e['a']]
                b = data['notes'][e['b']]
                ctx.move_to(
                    a['track'] * TRACKS_SCALE + TRACKS_GAP,
                    SURFACE_HEIGHT - (a['seconds'] * SECONDS_SCALE + SECONDS_GAP),
                )
                ctx.line_to(
                    b['track'] * TRACKS_SCALE + TRACKS_GAP,
                    SURFACE_HEIGHT - (b['seconds'] * SECONDS_SCALE + SECONDS_GAP),
                )
                ctx.stroke()

            for n in data['notes']:
                match n['kind']:
                    case 0:  # tap note
                        ctx.set_source_rgb(0, 1, 1)
                    case 1:  # drag note
                        ctx.set_source_rgb(1, 1, 0)
                    case 2:  # flick note
                        ctx.set_source_rgb(1, 0.5, 0.5)
                    case 3:  # throw note
                        ctx.set_source_rgb(0, 1, 0)
                    case 4:  # slide note
                        raise RuntimeError('?')
                ctx.move_to(
                    (n['track'] - n['width'] / 2) * TRACKS_SCALE + TRACKS_GAP + NOTE_GAP / 2,
                    SURFACE_HEIGHT - (n['seconds'] * SECONDS_SCALE + SECONDS_GAP) - NOTE_HEIGHT / 2,
                )
                ctx.line_to(
                    (n['track'] + n['width'] / 2) * TRACKS_SCALE + TRACKS_GAP - NOTE_GAP / 2,
                    SURFACE_HEIGHT - (n['seconds'] * SECONDS_SCALE + SECONDS_GAP) - NOTE_HEIGHT / 2,
                )
                ctx.line_to(
                    (n['track'] + n['width'] / 2) * TRACKS_SCALE + TRACKS_GAP - NOTE_GAP / 2,
                    SURFACE_HEIGHT - (n['seconds'] * SECONDS_SCALE + SECONDS_GAP) + NOTE_HEIGHT / 2,
                )
                ctx.line_to(
                    (n['track'] - n['width'] / 2) * TRACKS_SCALE + TRACKS_GAP + NOTE_GAP / 2,
                    SURFACE_HEIGHT - (n['seconds'] * SECONDS_SCALE + SECONDS_GAP) + NOTE_HEIGHT / 2,
                )
                ctx.close_path()
                ctx.fill()
