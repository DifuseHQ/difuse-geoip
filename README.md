# difuse-geoip

A simple go program to download and parse the MaxMind GeoIP2 country database (CSV) and organize it into a .zone file for each country. It also runs an HTTP server using fiber to serve the zone files as a gunziped tarball.

## Usage

### Building

```bash
go build
```

### Running

```bash
./difuse-geoip
```

Remember to set the LICENSE_KEY environment variable in .env to your MaxMind license key.