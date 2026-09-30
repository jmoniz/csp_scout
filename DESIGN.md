---
name: CSP Scout
description: A local CSP violation triage console where every rule traces back to observed evidence.
colors:
  deep-console: "oklch(12.9% 0.042 264.695)"
  console-surface: "oklch(20.8% 0.042 265.755)"
  panel-slate: "oklch(27.9% 0.041 260.031)"
  edge-slate: "oklch(37.2% 0.044 257.287)"
  edge-slate-strong: "oklch(44.6% 0.043 257.281)"
  text-bright: "oklch(96.8% 0.007 247.896)"
  text-body: "oklch(86.9% 0.022 252.894)"
  text-muted: "oklch(70.4% 0.04 256.788)"
  text-faint: "oklch(55.4% 0.046 257.417)"
  beacon-indigo: "oklch(51.1% 0.262 276.966)"
  beacon-indigo-bright: "oklch(58.5% 0.233 277.117)"
  beacon-indigo-text: "oklch(67.3% 0.182 276.935)"
  cleared-green: "oklch(76.5% 0.177 163.223)"
  cleared-green-deep: "oklch(69.6% 0.17 162.48)"
  unreviewed-amber: "oklch(82.8% 0.189 84.429)"
  blocked-rose: "oklch(71.2% 0.194 13.428)"
  first-party-sky: "oklch(74.6% 0.16 232.661)"
  wildcard-teal: "oklch(85.5% 0.138 181.071)"
typography:
  display:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.875rem"
    fontWeight: 700
    lineHeight: "2.25rem"
    letterSpacing: "-0.025em"
  headline:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 700
    lineHeight: "2rem"
    letterSpacing: "-0.025em"
  title:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 700
    lineHeight: "1.75rem"
    letterSpacing: "-0.025em"
  subtitle:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 700
    lineHeight: "1.25rem"
  body:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 400
    lineHeight: "1rem"
  label:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.6875rem"
    fontWeight: 600
    lineHeight: "1rem"
    letterSpacing: "0.05em"
  micro:
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.625rem"
    fontWeight: 500
    lineHeight: "0.875rem"
  code:
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace"
    fontSize: "0.75rem"
    fontWeight: 400
    lineHeight: "1.25rem"
rounded:
  sm: "0.25rem"
  md: "0.5rem"
  lg: "0.75rem"
  xl: "1rem"
  full: "9999px"
spacing:
  unit: "0.25rem"
  xs: "0.375rem"
  sm: "0.5rem"
  md: "0.75rem"
  lg: "1rem"
  xl: "1.25rem"
components:
  button-primary:
    backgroundColor: "{colors.beacon-indigo}"
    textColor: "oklch(100% 0 0)"
    rounded: "{rounded.md}"
    padding: "6px 12px"
  button-primary-hover:
    backgroundColor: "{colors.beacon-indigo-bright}"
    textColor: "oklch(100% 0 0)"
    rounded: "{rounded.md}"
    padding: "6px 12px"
  button-secondary:
    backgroundColor: "{colors.panel-slate}"
    textColor: "{colors.text-body}"
    rounded: "{rounded.md}"
    padding: "4px 8px"
  button-approve-self:
    backgroundColor: "oklch(58.8% 0.158 241.966)"
    textColor: "oklch(100% 0 0)"
    rounded: "{rounded.md}"
    padding: "6px 12px"
  button-danger:
    backgroundColor: "oklch(51.4% 0.222 16.935)"
    textColor: "oklch(100% 0 0)"
    rounded: "{rounded.md}"
    padding: "4px 8px"
  card-panel:
    backgroundColor: "{colors.console-surface}"
    rounded: "{rounded.lg}"
    padding: "20px"
  well-code:
    backgroundColor: "{colors.deep-console}"
    rounded: "{rounded.md}"
    padding: "16px"
  input-text:
    backgroundColor: "{colors.deep-console}"
    textColor: "{colors.text-bright}"
    typography: "{typography.code}"
    rounded: "{rounded.md}"
    padding: "6px 10px"
  pill-filter:
    backgroundColor: "{colors.console-surface}"
    textColor: "{colors.text-muted}"
    rounded: "{rounded.md}"
    padding: "6px 12px"
  pill-filter-active:
    backgroundColor: "{colors.beacon-indigo}"
    textColor: "oklch(100% 0 0)"
    rounded: "{rounded.md}"
    padding: "6px 12px"
  status-badge:
    textColor: "{colors.unreviewed-amber}"
    rounded: "{rounded.full}"
    padding: "2px 8px"
---

# Design System: CSP Scout

## Overview

**Creative North Star: "The Signal Room"**

CSP Scout is a dark room with one live indicator and a queue of unknown signals arriving in it. The visitor is an engineer on shift, not a shopper: the interface is an instrument they read, and everything in it exists to move an unknown origin from *pending* to *cleared* or *blocked* with the evidence visible at the moment of the decision.

The whole system is a single luminance ladder in one hue family. Nothing sits on paper; every surface is a step on a near-black slate ramp (`deep-console` at the bottom, `panel-slate` for controls), and the only saturated color in the room is a single indigo beacon reserved for the current position — the active filter, the primary action. Color appears otherwise only as a verb: green cleared it, amber has not seen it, rose refused it, sky recognized it as our own.

Density is deliberate and high. Body text is `0.75rem`, table headers are `0.6875rem` uppercase, and machine strings — origins, wildcards, directives, compiled policy — are monospace so they read as data and copy as data. The tool is quiet until a state changes, and then exactly one thing changes color.

**Key Characteristics:**
- One dark slate ladder, near-black to near-white, with no light mode.
- A single indigo accent used for the current position and the primary action, not for decoration.
- Semantic color as a verb set: cleared, unreviewed, blocked, first-party, wildcard.
- Monospace for every string the user will copy or compare.
- Very small, very dense type; the console is read, not skimmed.
- Depth from tonal layering and hairline borders; shadows stay near-flat.

## Colors

A monochrome slate instrument with one indigo beacon and five semantic signals. Every value is a Tailwind 4 default; there is no project `@theme` block, so these slugs are the names the codebase should use.

### Primary
- **Beacon Indigo** (`oklch(51.1% 0.262 276.966)`): the active filter pill, the primary button fill, the checkbox and focus ring. It marks current position and the one forward action per view.
- **Beacon Indigo Bright** (`oklch(58.5% 0.233 277.117)`): hover state of anything filled with Beacon Indigo, and the focus border on inputs and selects.
- **Beacon Indigo Text** (`oklch(67.3% 0.182 276.935)`): indigo as text — links, inline code references, the "view all processed rules" affordance, the header version badge.

### Secondary
- **First-Party Sky** (`oklch(74.6% 0.16 232.661)`): the `'self'` family. Marks an origin that matches the session target, the first-party filter segment, and the one-click "Approve All 'self'" action. Sky means *this resource is ours*.
- **Wildcard Teal** (`oklch(85.5% 0.138 181.071)`): the wildcard-approved state, and nothing else. It exists to keep a widened rule visually distinct from a widened-safety decision.

### Tertiary
- **Cleared Green** (`oklch(76.5% 0.177 163.223)`): approved states, the API-online dot, the inbox-zero state.
- **Cleared Green Deep** (`oklch(69.6% 0.17 162.48)`): fill-level green where a solid is needed (`bg-emerald-500`) — status dots and the online indicator.
- **Unreviewed Amber** (`oklch(82.8% 0.189 84.429)`): pending. The default state of everything newly captured, the "remaining" counter, report-only mode.
- **Blocked Rose** (`oklch(71.2% 0.194 13.428)`): rejected, error, API-offline. Also the only destructive color.

### Neutral
- **Deep Console** (`oklch(12.9% 0.042 264.695)`): the page floor, code wells, input fields, and the segmented-control track. The darkest step.
- **Console Surface** (`oklch(20.8% 0.042 265.755)`): cards, panels, stacked control bars. One step up from the floor.
- **Panel Slate** (`oklch(27.9% 0.041 260.031)`): the default border color (the most used value in the codebase) and raised control fills.
- **Edge Slate** (`oklch(37.2% 0.044 257.287)`): stronger borders on interactive controls; the default hover fill.
- **Edge Slate Strong** (`oklch(44.6% 0.043 257.281)`): the muted border used on the ignored state and low-emphasis outlines.
- **Text Bright** (`oklch(96.8% 0.007 247.896)`): headings, table values the user reads, monospace origins in the triage table.
- **Text Body** (`oklch(86.9% 0.022 252.894)`): control labels, table cells.
- **Text Muted** (`oklch(70.4% 0.04 256.788)`): the workhorse secondary text — descriptions, captions, inactive filters.
- **Text Faint** (`oklch(55.4% 0.046 257.417)`): timestamps, placeholders, empty-state prose, the lowest readable step.

### Named Rules
**The One Beacon Rule.** Beacon Indigo appears once per visual region as the active state or the single forward action. If two indigo fills compete in one panel, one of them is wrong.

**The Verb Rule.** Semantic colors are only verbs: sky = ours, teal = widened, green = cleared, amber = unseen, rose = refused. A color outside these meanings is drift, and the codebase currently carries some (see Don'ts).

## Typography

**Display Font:** the system UI stack (`ui-sans-serif, system-ui, sans-serif`) — no webfont is loaded.
**Body Font:** the same system stack.
**Label/Mono Font:** `ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas` for all machine strings.

**Character:** a developer-tool pairing — the OS's own UI face for prose, the OS's own monospace for data. Nothing is downloaded, nothing is decorative. The result reads like an instrument panel: the human layer is small and neutral, the machine layer is unmistakable.

### Hierarchy
- **Display** (700, `1.875rem`/`2.25rem`, tracking `-0.025em`): the list-view page title. One per screen.
- **Headline** (700, `1.5rem`/`2rem`, tracking `-0.025em`): section headings and the four metric numbers (`text-2xl`). The metric numbers are the only place large type carries data.
- **Title** (700, `1.25rem`/`1.75rem`, tracking `-0.025em`): the session name on the detail view.
- **Subtitle** (700, `0.875rem`/`1.25rem`): card and panel titles ("Live Compiled CSP", "Recent Violation Samples").
- **Body** (400–500, `0.75rem`/`1rem`): the working size for nearly everything — table cells, descriptions, controls. This is the single most used size in the system by a wide margin.
- **Label** (600, `0.6875rem`/`1rem`, tracking `0.05em`, uppercase): table column headers only.
- **Micro** (500, `0.625rem`/`0.875rem`): status badges, the `'self' match` tag, the version badge. Non-interactive only.
- **Code** (400, `0.75rem`/`1.25rem`, monospace): origins, wildcards, directives, source files, line numbers, and the compiled policy output.

### Named Rules
**The Monospace Is Data Rule.** Any string the user will copy, compare against a header, or paste into a server config is monospace. Origins, wildcards, directives and the compiled policy are never set in the UI face.

**The No Download Rule.** The system stack is the type system. A webfont is a dependency and a load; if one is ever added it must earn its place against the instrument-panel character.

## Layout

A single centered column with a sticky top bar. The container is `max-w-7xl` with responsive gutters (`1rem` mobile, `1.5rem` tablet, `2rem` desktop), and the top bar is a fixed `4rem` tall row with a translucent background and backdrop blur.

The detail view is a vertical stack of regions: metric strip, the policy panel, then the triage surface. The triage surface is the only wide element and is a real HTML table inside a horizontal scroll container — it is designed for a laptop, not a phone, and it overflows rather than reflowing.

Spacing is Tailwind's `0.25rem` unit used at a small set of ad-hoc steps (`0.375rem`–`1.25rem` for control gaps, `0.5rem`–`3rem` for region padding). Density inside a card is tight (`0.5rem`–`0.75rem` padding on table cells); density between regions is generous enough to separate them without whitespace becoming decoration.

## Elevation & Depth

Tonal layering is the primary depth mechanism, and it is strict: the floor is `deep-console`, cards are `console-surface`, controls inside cards are `panel-slate` or return to `deep-console` as wells. A nested surface never goes lighter than `panel-slate`.

Shadows are near-flat and mostly a side effect of the Tailwind defaults in use (`shadow`, `shadow-sm`), plus `shadow-xl` on the policy panel and `shadow-2xl` on modal dialogs. There are exactly two purposeful exceptions: the indigo-tinted glow under the primary CTA (`shadow-indigo-500/20`) and the two hardcoded status-dot glows (`0 0 8px rgba(...)` on the API online/offline dots). Hover on a primary action lifts by `1px`–`2px` with a scale nudge; nothing else moves.

### Named Rules
**The Border-First Rule.** Separation is drawn with a hairline `panel-slate` border before it is drawn with a shadow. Depth is tonal; borders are structural.

**The Still Room Rule.** The room is still. Motion is limited to state feedback under `200ms`, plus the two status-dot pulses. Nothing animates on mount except modals fading in.

## Shapes

A single soft-corner language with four steps in active use, plus full round for badges:

- `0.25rem` — inline chips, code tokens, the directive badge, checkbox corners.
- `0.5rem` — the default control radius: buttons, inputs, selects, filter pills, status dots' container.
- `0.75rem` — cards, panel sections, the table container, the code well.
- `1rem` — the two most prominent containers only: the policy panel and modal dialogs.
- `9999px` — status badges and filter pill dots, never a container.

Borders are always `1px` and always from the slate ladder; there are no thick borders, no double borders, and no dashed borders except the single empty-state container, which is dashed on purpose to read as absence rather than a surface.

## Components

### Buttons
- **Shape:** `0.5rem` radius, `1px` `panel-slate` border, `0.375rem`–`0.75rem` padding, `0.75rem` semibold text.
- **Primary:** Beacon Indigo fill, white text, `6px 12px` padding, subtle indigo glow shadow; on hover the fill moves to Beacon Indigo Bright and the button lifts `1px`.
- **Secondary / table actions:** `panel-slate` fill with `edge-slate` border and Text Body label. Their hover treatment recolors toward the semantic hue of the action they perform (approve → green, wildcard → teal, `'self'` → sky) rather than to indigo. This is the signature interaction of the triage table: the button tells you the verb before you click it.
- **Destructive:** Sheet-level fills only for confirm actions (`rose-700`), otherwise a quiet icon-only button that turns Blocked Rose on hover. Destructive actions are never the primary fill in a region.
- **Approve All 'self':** the one solid sky button in the system, present only while first-party pending items exist. Its rarity is the point.

### Segmented Control
- **Style:** a `deep-console` track with `0.25rem` inner padding and a `panel-slate` border, holding two or three equal options.
- **Active:** solid fill and white text — indigo for the default position, sky when the active segment is the first-party filter.
- **Use:** mutually exclusive view modes only (source type, output format). Never for independent toggles.

### Filter Pills
- **Style:** `0.5rem` radius, `0.75rem` medium text, horizontally scrollable single row.
- **Active:** Beacon Indigo fill, white text, `shadow-sm`.
- **Inactive:** `console-surface` fill, `panel-slate` border, Text Muted label, hovering to white text on `edge-slate`.

### Status Badges
- **Style:** full-round, `0.625rem` medium text, `2px 8px` padding, and the consistent tint recipe: 10–20% background of the semantic hue, a 30% border of the same hue, and the 300-step text of that hue.
- **States:** Pending Triage (amber), Origin Approved (green), Wildcard Approved (teal), `'self'` Approved (sky), Rejected (rose), Ignored (neutral slate).
- **Rule:** the badge always carries its label as text. Color is the reinforcement, never the message.

### Directive Chips
- **Style:** `0.25rem` radius, `0.6875rem` monospace, bordered, and hue-keyed per directive family: `connect-src` cyan, `script-src` amber, `style-src` fuchsia, `img-src` green, `font-src` indigo, `frame-src` rose, everything else neutral slate.
- **Note:** this is the one place where hue is used as a category key rather than a state, and it is the loudest drift in the system — six hues for taxonomy sit next to five hues for meaning.

### Cards / Containers
- **Corner Style:** `0.75rem` for panels, `1rem` for the two headline containers.
- **Background:** `console-surface` at 60–80% opacity over the floor.
- **Border:** `1px` `panel-slate`, always.
- **Shadow Strategy:** `shadow`/`shadow-sm` by default; `shadow-xl` reserved for the policy panel.
- **Internal Padding:** `0.75rem` for control bars, `1.25rem` for panels.

### Inputs / Fields
- **Style:** `deep-console` well, `1px` `panel-slate` border, `0.5rem` radius, `0.75rem` text (monospace for directive values), Text Bright value, Text Faint placeholder.
- **Focus:** the border moves to Beacon Indigo Bright. There is no outline, no ring on text inputs, and no fill change.
- **Labels:** `0.75rem` semibold Text Body, always above the field, always with a `for`/`id` pair.

### Navigation
- **Style:** sticky `4rem` bar, `console-surface` at 80% with backdrop blur, `1px` bottom border.
- **Contents:** the product mark and version badge on the left, the live API status indicator and the Sessions link on the right. The version badge is indigo-tinted micro type.
- **Live status:** a colored dot plus a text label ("API Online"/"API Offline"/"Checking…"), clickable to force a re-check. It is the room's heartbeat and the only element with a pulsing animation.

### Triage Table
- **Character:** the signature surface. A dark table whose rows are tasks, not records.
- **Header:** `0.6875rem` uppercase tracking-wide Text Muted on a `deep-console` band.
- **Rows:** hairline `divide-slate-800` separators, 30% slate hover, and an indigo-tinted background when selected.
- **Columns:** directive chip, blocked origin (monospace, selectable) with a `'self' match` micro tag, wildcard suggestion in a code well, report count as the expand affordance, status badge, inline triage actions.
- **Evidence drawer:** expanding a row opens a full-width sub-table in a `deep-console` band with monospace sample data (document URI, blocked URI, source script, line:column, time).

### Code Well
- **Style:** `deep-console` floor, `panel-slate` border, `0.75rem` radius, `1rem` padding, monospace, wrapped and selectable.
- **Default text color:** Cleared Green for the compiled policy output.

## Do's and Don'ts

### Do:
- **Do** keep every surface on the slate ladder by role: `deep-console` for wells and the floor, `console-surface` for cards, `panel-slate` for controls and borders, `edge-slate` for interactive borders and hover.
- **Do** set every copyable machine string in `{typography.code}` — origins, wildcards, directives, hashes, line numbers, policy output.
- **Do** use the badge recipe: 10–20% tint background, 30% same-hue border, 300-step same-hue text, full-round, `0.625rem`.
- **Do** keep `{typography.body}` (`0.75rem`) as the working size. Density is the point of this console.
- **Do** let table-action buttons announce their verb on hover by recoloring toward the semantic hue they will apply.
- **Do** pair every semantic color with its label text.
- **Do** separate with a hairline `panel-slate` border first; reach for a shadow only for modals and the policy panel.

### Don't:
- **Don't** add a fifth radius step. Four are in use plus full-round; new surfaces pick from those.
- **Don't** use Beacon Indigo as decoration, as a background wash, or for two competing actions in the same region.
- **Don't** render the compiled policy in Cleared Green. It collides with the approved state and makes the output look like a status. Use Text Bright or a neutral step.
- **Don't** introduce a Tailwind hue outside the assigned verbs. `fuchsia`, `cyan` and `blue` are currently drifting in the directive chips and header gradient; they need roles or removal.
- **Don't** use a webfont, an icon font, or an icon package. Icons are inline SVG with `stroke-width: 2`.
- **Don't** animate on mount, on scroll, or on hover beyond a `1px`–`2px` lift; nothing in the system exceeds `300ms` except the two status-dot pulses.
- **Don't** convey triage state with color alone — the badge label is mandatory.
- **Don't** put a shadow on a table row, a filter pill, or an input.
