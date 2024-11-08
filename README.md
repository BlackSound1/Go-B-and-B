# Go B & B

A web app for a fictional Bed and Breakfast, written in Go.

![Home Page](images/Home%20Page.png)

## Features

- Can book stays to 2 rooms for any length of time.
- Email confirmations for owner and guests.
- Admin dashboard hidden behind Auth.
  - Admin can process new reservations.
  - Admin can cancel new reservations.
  - Admin can block off days when a room is not available.
  - Admin can see all reservations.
  - Admin can see new, unprocessed reservations.
  - Admin can see monthly calendar of reservations.
  - Log in/ out functionality.

## Tech Stack

- Go: 1.23.1
- Database: PostgreSQL with [PGX](https://github.com/jackc/pgx)
- Email: [MailHog](https://github.com/mailhog/MailHog) and [Go Simple Mail](https://github.com/xhit/go-simple-mail)
- `.env ` Management: [Go Dotenv](https://github.com/joho/godotenv)
- Form Validation: [Go Validator](https://github.com/asaskevich/govalidator)
- CSRF Prevention: [NoSurf](https://github.com/justinas/nosurf)
- HTTP Routing: [Chi Router](https://github.com/go-chi/chi)
- Session Management: [SCS](https://github.com/alexedwards/scs/)
- Database Migrations: [Pop](https://gobuffalo.io/documentation/database/pop/)/ [Soda](https://gobuffalo.io/documentation/database/soda/)
- Admin Dashboard: [Royal UI Free Bootstrap Admin Template](https://github.com/BootstrapDash/RoyalUI-Free-Bootstrap-Admin-Template)
- Frontend: Bootstrap
- Notifications: [Notie](https://github.com/jaredreich/notie)
- Alerts: [SweetAlert 2](https://sweetalert2.github.io/)
- Datepickers: [VanillaJS Datepicker](https://github.com/mymth/vanillajs-datepicker)
- Building: Make
- Cloud: Linode/ Akamai
- Cloud Webserver: [Caddy](https://caddyserver.com/docs/install#debian-ubuntu-raspbian)

## Server Setup

This project uses Linode/ Akamai for its server.

1. Log in as root, connecting by the Linode IP address: `ssh root@<IP ADDRESS>`
2. Update server:
   ```sh
   apt update && apt upgrade -y
   ```
3. Install Caddy:
   ```sh
   apt install -y debian-keyring debian-archive-keyring apt-transport-https curl && \
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg && \
   curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list && \
   apt update && apt install caddy
   ```
4. Install supervisor:
   ```sh
   apt install supervisor
   ```
5. Install PostgreSQL (version 16 is used here. May not be available when reading this):
   ```sh
   apt install postgres-##
   ```
6. Create new user (accept defaults):
   ```sh
   adduser <NAME>
   ```
7. Give user root permissions:
   ```sh
   usermod -aG sudo <NAME>
   ```
8.  Log in as user: `<NAME>@<IP ADDRESS>`
9. Install Make:
   ```sh
   sudo apt install make
   ```
10. Download Go: 
   ```sh
   wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz
   ```
11. Install Go: 
    ```sh
    sudo tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz
    ```
12. Add Go to PATH:
    ```sh
    export PATH=$PATH:/usr/local/go/bin
    ```
13. To make sure this Go is always used for this user, update `.profile` with the above `export`
14. Clone this repo: `git clone https://github.com/BlackSound1/Go-B-and-B.git`
15. Adjust Postgres configuration located at `/etc/postgresql/##/main/pg_hba.conf`. Adjust IPv4 and IPv6 `METHOD` to 'trust' to keep it simple.
16. Restart Postgres:
    ```sh
    sudo service postgresql stop && sudo service postgresql start
    ```
17. Populate the `database.yml` file from the `database.yml.example` file and fill out the fieleds as appropriate.
    ```sh
    cp database.yml.example database.yml
    ```
18. Get Pop: `go install github.com/gobuffalo/pop/v6/soda@latest`
29. Add Soda to PATH by editing `.profile` to add `export PATH=$PATH:~/go/bin`
20. Run migrations: `soda migrate`
21. Populate the `.env` file from the `.env.example` file and fill out the fields as approprate (setting `PROD` to `true`).
   ```sh
   cp .env.example .env
   ```
22. Build and run the app: `make build`

## How to Run

If you have Make installed: `make run`.

If you don't: `go run ./cmd/web` or `go run ./...`.

## How to Log in as Admin

Hit the "Log In" button on top. The admin username is: "me@me.me" and the password is: "cool_dude".
These credentials can be exposed like this because the app is a proof-of-concept for portfolio purposes only.
