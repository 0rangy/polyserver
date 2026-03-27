# Polyserver

## About
Polyserver is a headless server for PolyTrack

## Usage
### Download
Grab the binary for your system from https://github.com/polytrackmods/polyserver-go/releases
If the right binary for your system is not there, you can also build the server yourself by cloning the repository and running `go build .`

### Tracks
The server looks for tracks inside the `tracks` folder inside the working directory in the form of `.track` files. `.track` files are just raw text files containing a track code

You can grab the .track files for the official and community tracks from the releases page

In the end, the folder structure should look something like this:
```
|--- tracks
    |--- community
        |--- 4_seasons.track
        |--- 90_reset.track
        |--- anubis.track
        |--- arabica.track
        |--- ...
    |--- custom
        |--- test1.track
    |--- official
        |--- desert1.track
        |--- desert2.track
        |--- ...
|--- polyserver.exe
```

### Running the Server
To run the server, simply run the binary you just downloaded
#### Launch options
`port`: the port the server management dashboard runs on. Default is `8080`.  
`control-port`: the port the control server runs on. Default is `9090`.  
`address`: The address the web frontend binds to. If you want the frontend to only be available to the host, either ignore this or set it to `127.0.0.1`. If you want the frontend to bind to all addresses, set this to `0.0.0.0`. Default is `127.0.0.1`.