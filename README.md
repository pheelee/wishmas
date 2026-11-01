![Wishmas — Christmas lights and snowflakes](docs/screenshots/hero.svg)

# Wishmas

[![Latest release](https://img.shields.io/github/v/release/pheelee/wishmas)](https://github.com/pheelee/wishmas/releases/latest)
[![Go statement coverage](https://raw.githubusercontent.com/pheelee/wishmas/badges/coverage.svg)](https://github.com/pheelee/wishmas/blob/badges/coverage.md)
[![GHCR downloads](https://ghcr-badge.elias.eu.org/shield/pheelee/wishmas/wishmas)](https://github.com/pheelee/wishmas/pkgs/container/wishmas)

A self-hosted family wishlist app with a dark December-night theme, colorful Christmas lights and gentle snowfall. Built with Go, Tailwind CSS and vanilla JavaScript.

## Screenshots

An example household with fictional wishes and gift-givers. Select an image to see it at full size.

### Wishes and claims

Pictures, reference prices, shop links and colored wish states. Gift-givers recognize their own claims without seeing other claimants’ names.

[![Wishlist with pictures and claim states](docs/screenshots/wishlist-desktop.png)](docs/screenshots/wishlist-desktop.png)

<table>
  <tr>
    <th>Setup and appearance</th>
    <th>Edit a wish</th>
  </tr>
  <tr>
    <td valign="top"><a href="docs/screenshots/setup-desktop.png"><img src="docs/screenshots/setup-desktop.png" alt="Setup wizard with household, language and Christmas appearance controls" width="460"></a></td>
    <td valign="top"><a href="docs/screenshots/edit-wish.png"><img src="docs/screenshots/edit-wish.png" alt="Wish editing modal with picture, price, links and note" width="460"></a></td>
  </tr>
</table>

### Manage recipients

Parents choose recipients and see reference totals and claim counts for each list.

[![List recipient management and totals](docs/screenshots/manage-recipients.png)](docs/screenshots/manage-recipients.png)

<table>
  <tr>
    <th>Mobile overview</th>
    <th>Mobile wishlist</th>
  </tr>
  <tr>
    <td valign="top"><a href="docs/screenshots/overview-mobile.png"><img src="docs/screenshots/overview-mobile.png" alt="Mobile recipient overview with totals and claim badges" width="340"></a></td>
    <td valign="top"><a href="docs/screenshots/wishlist-mobile.png"><img src="docs/screenshots/wishlist-mobile.png" alt="Mobile wish cards with shop links and claim actions" width="340"></a></td>
  </tr>
</table>

## Features

- Create lists for different occasions and choose their recipients from your household.
- Keep wishes in a shared household pool, then assign them to a recipient and list.
- Add optional pictures, multiple shop links, a reference price and notes.
- Share individual links with gift-givers. They can claim wishes without creating accounts; parents manage everything with a shared password.
- See recipient totals and claimed/unclaimed counts. Gift-givers see their own claims and claim counts; parents can see all claimant names.
- Use English or Swiss German, with a responsive layout for desktop and mobile.
- Store household data in YAML locally or in S3-compatible storage. Pictures stay on disk.

## Run with Docker Compose

Use an existing Traefik installation configured for HTTPS and HTTP-to-HTTPS redirection. Wishmas runs behind the proxy; its backend port stays private.

Create a `.env` file with your domain, existing Traefik network, certificate resolver and the actual Traefik peer IP or restricted CIDR. Replace these example values:

```dotenv
WISHMAS_HOST=wishmas.example.com
TRAEFIK_NETWORK=your-traefik-network
TRAEFIK_CERT_RESOLVER=your-certificate-resolver
WISHMAS_TRUSTED_PROXIES=your-traefik-ip/32
```

Save this as `compose.yaml`:

```yaml
services:
  wishmas:
    image: ghcr.io/pheelee/wishmas:latest
    restart: unless-stopped
    environment:
      WISHMAS_TRUSTED_PROXIES: "${WISHMAS_TRUSTED_PROXIES:?Set WISHMAS_TRUSTED_PROXIES}"
    volumes:
      - wishmas-data:/data
    networks:
      - proxy
    labels:
      traefik.enable: "true"
      traefik.docker.network: "${TRAEFIK_NETWORK:?Set TRAEFIK_NETWORK}"
      traefik.http.routers.wishmas.rule: "Host(`${WISHMAS_HOST:?Set WISHMAS_HOST}`)"
      traefik.http.routers.wishmas.entrypoints: "${TRAEFIK_ENTRYPOINT:-websecure}"
      traefik.http.routers.wishmas.tls: "true"
      traefik.http.routers.wishmas.tls.certresolver: "${TRAEFIK_CERT_RESOLVER:?Set TRAEFIK_CERT_RESOLVER}"
      traefik.http.services.wishmas.loadbalancer.server.port: "8080"

volumes:
  wishmas-data:

networks:
  proxy:
    external: true
    name: "${TRAEFIK_NETWORK:?Set TRAEFIK_NETWORK}"
```

If your HTTPS entrypoint has another name, set `TRAEFIK_ENTRYPOINT` in `.env`. Run one Wishmas instance per household data volume.

```sh
docker compose up -d
```

Open `https://your-domain` and complete the setup wizard to choose your household name, admin password, language and appearance. Then add recipients, gift-givers, lists and wishes.

The image runs as UID/GID `1000:1000`. Keep `/data` across container replacements, including when using S3. For custom UID/GID or bind mounts, use the [Compose reference](docker-compose.yml) and make the mounted directory writable by the selected user. See [S3 configuration](docs/s3-storage.md) and [password administration](docs/offline-admin.md) for those options.

### Build from source

Clone this repository and build the image locally before starting the same Compose configuration:

```sh
git clone https://github.com/pheelee/wishmas.git
cd wishmas
docker build -t ghcr.io/pheelee/wishmas:latest .
```
