# Luna (fork)

> **Fork notice — differences from upstream (as of 2026-06-26)**
>
> This is a fork of [Opisek/luna](https://github.com/Opisek/luna). Changes made in this fork that are not (yet) in upstream:
>
> - **Fix: invalid `TZID:Local` exported for events with no explicit timezone**, causing Android/DAVx5 to reject sync with `unknown timezone id: Local`. `ParseIcalTime` now falls back to `UTC` instead of the literal string `"Local"` when no `TZID`/`Z` is present. All-day detection in `ParseIcalEvent` was reworked to check the raw `DATE` vs `DATE-TIME` property format instead of relying on the now-removed `time.Local` sentinel. See [backend/src/types/date.go](backend/src/types/date.go) and [backend/src/protocols/internal/common.go](backend/src/protocols/internal/common.go).
> - **Add CI workflow to build and push Docker images** (`lunafork-backend`, `lunafork-frontend`, `linux/amd64` only) to Docker Hub on every branch push. See [.github/workflows/docker-build-push.yml](.github/workflows/docker-build-push.yml).
> - **Add "Copy to dates..." event duplication feature.** From an existing event's view, duplicate it to one or more non-contiguous dates picked from a navigable mini-calendar popup, preserving name/description/color/duration/all-day flag; copies are always plain non-recurring events. See [frontend/src/components/modals/CopyEventModal.svelte](frontend/src/components/modals/CopyEventModal.svelte), the "Copy to dates..." button and `buildEventCopy`/`onCopyToDates` in [frontend/src/components/modals/EventModal.svelte](frontend/src/components/modals/EventModal.svelte), and the `isSelected`/`dynamicRows` props added to [frontend/src/components/interactive/SmallCalendar.svelte](frontend/src/components/interactive/SmallCalendar.svelte).
> - **Add drag-and-drop event rescheduling** on the calendar grid (day-cell granularity, native HTML5 DnD). Dragging an event onto another day shifts its start/end by whole days, preserving time-of-day and duration. See `draggable`/`ondragstart` in [frontend/src/components/calendar/Event.svelte](frontend/src/components/calendar/Event.svelte) and the drop handlers in [frontend/src/components/calendar/Day.svelte](frontend/src/components/calendar/Day.svelte). This also fixes a latent bug where `Repository.editEvent` never refreshed the reactive `events` array for events staying within the visible date range, and extracts a shared `Repository.getEventSourceType` helper — see [frontend/src/lib/client/data/repository.svelte.ts](frontend/src/lib/client/data/repository.svelte.ts).
> - **Improve the "today" cell highlight.** Previously only a small circle around the day number; now the entire day cell gets a border in `$foregroundPrimary` (the active theme's main text color, e.g. white on dark themes, black on light themes) for much better visibility regardless of theme. See [frontend/src/components/calendar/Day.svelte](frontend/src/components/calendar/Day.svelte).

📅 Luna is a self-hosted **calendar frontend** and **aggregator**.
- A single web-app for all of your calendars
- Differerent calendar protocols like **CalDav**, **iCal**, and **Google Calendar**
- User management

![Screenshot of luna in light and dark mode](./documentation/pictures/light-dark.png)

🎨 Luna is **infinitely customizable**, so your calendar can be as unique as you!
- Completely customizable themes and fonts
- Many built-in themes of popular color schemes
- Simple installation of additional themes and fonts
- Many ways to customize the look of the calendar

![Screenshot of luna with different themes](./documentation/pictures/themes.png)

# Disclaimer
Luna is an ambitious and large project. As such, development takes a long time.

At this point in time, Luna is nearing a usable 1.0.0 state, however, a lot of small and less small finishing touches still need to be done (in particular, full support for recurring events and mobile screen size support).

Due to my busy schedule and being the sole developer, I am unable to provide a release date for 1.0.0 for now. Feel free to get in contact through GitHub discussions if you are interested in contributing.

You may also follow the progress in the [development roadmap](https://todo.opisek.net/share/dvEazOyRLEYThqxohVosnqKskYLyoZ4nS8rQ63G1/auth?view=280).

# Getting Started

For instruction on how to deploy Luna, see [Deployment Guide](./documentation/deployment.md).

For a list of security mechanisms and compromise analyses, see [Security & Privacy](./documentation/security.md)

For a list of API endpoints, see [API](./documentation/api.md)
