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

## How to Run

1. Populate the `.env` file from the `.env.example` file and fill out the fields as approprate.
   ```sh
   cp .env.example .env
   ```
2. Populate the `database.yml` file fro mthe `database.yml.example` file and fill out the fieleds as appropriate.
    ```sh
    cp database.yml.example database.yml
    ```
3. Run all migrations:
   ```sh
   soda migrate
   ```
4. Run the app.
   1. If you have Make installed: `make run`
   2. If you don't: `go run ./cmd/web` or `go run ./...`

## How to Log in as Admin

Hit the "Log In" button on top. The admin username is: "me@me.me" and the password is: "cool_dude".
These credentials can be exposed like this because the app is a proof-of-concept for portfolio purposes only.
