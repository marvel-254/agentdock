# OrionOS Design System Master

**Project:** OrionOS
**Positioning:** Local Agent Control Plane
**Aesthetic:** Developer Mission Control / Technical / Premium

---

## Color Palette (Tokyo Night Inspired)

| Role | Hex | Usage |
|------|-----|-------|
| Background | `#0f0f14` | Deep background |
| Surface | `#16161e` | Cards, panels, sidebar |
| Border | `#1e1e2e` | Dividers, card strokes |
| Primary | `#7aa2f7` | Active states, focus, links |
| Success | `#9ece6a` | Working, Online, healthy |
| Warning | `#e0af68` | Waiting, Degraded |
| Danger | `#f7768e` | Blocked, Offline, Failed |
| Muted | `#565f89` | Labels, timestamps, inactive |
| Text | `#e9e9f0` | Body text |
| Secondary | `#a9b1d6` | Subheadings, sidebar items |

---

## Typography

- **Headings:** `Fira Code` (Monospace, bold)
- **Body:** `Fira Sans` (Sans-serif)
- **Technical Data:** `Fira Code` (Monospace)

---

## Layout Rules

- **Sidebar:** Fixed 240px width on desktop, hidden on mobile.
- **Top Bar:** 64px height, sticky.
- **Grid:** Sparse layout, cards with 8px radius, 1px border.
- **Mobile:** 60px bottom navigation, stacked vertical cards.

---

## Status Semantics

- **Working:** Active processing (Success green)
- **Active:** Running, minimal load (Primary blue)
- **Waiting:** Idle (Warning yellow)
- **Offline:** Stopped/Missing (Danger red)
