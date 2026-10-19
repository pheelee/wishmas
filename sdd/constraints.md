# Constraints

## CON-* entries

### CON-001: Hosting

- Deploy as a Docker container, self-hosted on a VPS.

### CON-002: Backend

- Use Go for the backend.

### CON-003: Frontend

- Use Tailwind CSS and vanilla JavaScript.
- The frontend has a Christmas theme.

### CON-004: Persistence

- Persist data in YAML files.
- Organize data hierarchically by household, then year, then event type, then person name. A person may be an adult or a child.
- Retain the event-type level for future expansion; Christmas is the only supported event in v1.
- Support managing household members and wishlists both through the app and by editing YAML files on the VPS.
- Pick up direct YAML file edits on page load.
- If manually edited YAML is invalid or contains duplicate person names, show an error and leave the files untouched.
- Block editing previous years' wishlists through the web UI. Direct YAML edits to those lists cannot be prevented and are an accepted risk.

## OPEN questions

