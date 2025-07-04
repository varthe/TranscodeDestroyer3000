> [!CAUTION]
> This is an experiment and is being tested. Use at your own discretion.

This proxy sits in front of plex like this:

`Internet -> Caddy/Nginx -> FMQ Proxy -> Plex`

It will intercept playback requests and force maximum quality. Other requests are unaffected.

### Setup

1. **Clone the repo**
   ```bash
   git clone https://github.com/varthe/TranscodeDestroyer3000.git
   ```
2. **Build the image**

   ```bash
   cd TranscodeDestroyer3000
   docker build -t fmq .
   ```

3. **Run with Docker Compose**  
   Ensure it's on the same network as Plex and your reverse proxy:

   ```yaml
   services:
     fmq:
       image: fmq
       container_name: fmq
       ports:
         - "8080:80"
       environment:
         - PLEX_URL=http://plex:32400 # Replace with your actual Plex URL
         - FORCE_MAXIMUM_QUALITY=true
         - DEBUG=true
       volumes:
         - ./logs:/proxy/logs
       restart: unless-stopped
       networks:
         - plex_network

   networks:
     plex_network:
   ```

4. **Point your reverse proxy (Caddy/Nginx) to `http://fmq:80`**
