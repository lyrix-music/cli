import dbus
import time
import requests


session_bus = dbus.SessionBus()
elisa_bus = session_bus.get_object(
    "org.mpris.MediaPlayer2.elisa", "/org/mpris/MediaPlayer2"
)


last_song = ""
last_artist = ""
was_playing = False


def get_song_and_post():
    global last_artist
    global last_song
    global was_playing

    spotify_properties = dbus.Interface(elisa_bus, "org.freedesktop.DBus.Properties")
    playback_status = spotify_properties.Get(
        "org.mpris.MediaPlayer2.Player", "PlaybackStatus"
    )

    metadata = spotify_properties.Get("org.mpris.MediaPlayer2.Player", "Metadata")

    current_song = metadata["xesam:title"]
    current_artist = metadata["xesam:artist"][0]

    if was_playing and playback_status != "Playing":
        last_song = None
        last_artist = None
        was_playing = False
        print(f"Song play has stopped")
        requests.post("http://127.0.0.1:5555/api/currentsong/1265317047", json={})

    if playback_status == "Playing" and (
        current_artist != last_artist or current_song != last_song
    ):
        was_playing = True
        last_song = current_song
        last_artist = current_artist
        print(f"Currently playing {last_song} by {last_artist}.")
        requests.post(
            "http://127.0.0.1:5555/api/currentsong/1265317047",
            json={"song": last_song, "artist": last_artist},
        )


def main():
    try:
        while True:
            get_song_and_post()
            time.sleep(5)
    except KeyboardInterrupt:
        requests.post("http://127.0.0.1:5555/api/currentsong/1265317047", json={})


if __name__ == "__main__":
    main()
