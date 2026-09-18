![Logo](docs/logo-light.svg)

This is a webapp that holds and presents images, for drawing reference. This repo holds the backend Go app.

See the [frontend repo](https://github.com/DanielOaks/drawref-frontend) for the interface.

## Environment variables

[.env.sample](./.env.sample) includes descriptions of all the environment variables you can use to configure the backend. These environment variables also work on the docker image.

## Docker quick start

You can get the backend up and running quickly with Docker. \
The `.env` file includes sample docker calls you can use to set up the dependencies:

- Postgres
- Minio (or S3)

Once you have those set up:

```bash
# in some specific directory, create and edit the .env file
curl -o .env https://raw.githubusercontent.com/DanielOaks/drawref-backend/release/.env.sample
vim .env

# start the backend
docker run -it --env-file .env -p 3300:3300 ghcr.io/danieloaks/drawref-backend:release

# in another terminal, start the frontend
docker run -it -e REACT_APP_DRAWREF_API=http://localhost:3300/api/ -e REACT_APP_DRAWREF_UPLOAD=http://localhost:3300/upload/ -p 3000:3000 ghcr.io/danieloaks/drawref-frontend:main
```

Finally, access the app on port 3000. If running the command locally, at http://localhost:3000

## Development

If you want to get started developing the app, you can use the below commands.

### Development Server

Start the development server on http://localhost:3300

```bash
go run .
```

### Production

Build the application for production:

```bash
make
```

## Licensing

The software itself is distributed under the ISC license. The sample data (in the `/samples` folder) is distributed under different terms as described in [the LICENSE file](./LICENSE.md).
