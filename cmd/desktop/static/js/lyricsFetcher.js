let currentTrack = ""
let currentArtist = ""
let currentPlayers = []


$(document).ready(function() {

    // Check for click events on the navbar burger icon
    $(".navbar-burger").click(function() {

        // Toggle the "is-active" class on both the "navbar-burger" and the "navbar-menu"
        $(".navbar-burger").toggleClass("is-active");
        $(".navbar-menu").toggleClass("is-active");
    });


    console.log("Restoring old values")
    if (window.localStorage.getItem("scrobble") === "true") {
        // restore the last scrobbled track
        console.log("Restoring scrobble");
        $("#scrobbleSwitch").prop("checked", true)
    }

    let lastChosenPlayer = window.localStorage.getItem("lastChosenPlayer")
    if (lastChosenPlayer !== null && lastChosenPlayer !== "") {
        console.log("Restoring last player: " + lastChosenPlayer)
        let playerName = lastChosenPlayer.replace("org.mpris.MediaPlayer2.", "")
        $("#selected-music-player").text(capitalizeFirstLetter(playerName))
        $(".dropdown").removeClass("is-active")

    }

    $.get("/api/v1/user/logged-in", function(data) {
        console.log("Received data from /user/logged-in")
        if (data["logged_in"] === false) {
            console.log("User is not logged in, disabling scrobble switch")
            $(".requires-auth").remove()
        } else {
            $("#navBarLoginButton").text("Logout")
        }
    }, "json")


    function playerChangedCallback(item) {
        console.log("Requesting server to change player")
        let data = {
            "dbus_id": item.text
        }
        $.postJSON("/api/v1/player", data, function () {
            console.log(`Player changed successfully, received ${data} as player name`)
        }, { dataType: "text"})
        $(".dropdown").removeClass("is-active")
        window.localStorage.setItem("lastChosenPlayer", item.text)
        let playerName = item.text.replace("org.mpris.MediaPlayer2.", "")
        $("#selected-music-player").text(capitalizeFirstLetter(playerName))
    }

    function setScrobble(enabled) {
        console.log("Requesting server to change scrobble settings")
        $.post(`/api/v1/prefs/scrobble/${enabled}`, "", function () {
            console.log("Player changed successfully")
        }, "text")
        window.localStorage.setItem("scrobble", enabled)
    }


    function getLyrics() {
        console.log("Trying to fetch Lyrics")
        $.get("/api/v1/updates/lyrics", function (data) {
            console.log("Received lyrics")
            $(".lyrics").html(data.replaceAll("\n", "<br>"))
        }, "text")

    }

    function getPlayers() {
        $.get("/api/v1/updates/players", function (data) {
            if (data === currentPlayers) {
                return
            }
            // the players changed
            currentPlayers = data
            let dropdown = $(".dropdown-content")
            dropdown.empty()
            currentPlayers.forEach(function (item) {
                dropdown.append(`<a class="dropdown-item">${item}</a>`)
            })
            $(".dropdown-item").click(function () { playerChangedCallback(this) })

        }, "json")
    }


    function getUpdates() {
        console.log("Trying to get updates on song")

        $.get("/api/v1/updates/song", function (data) {
            // console.log("Received updates")
            let track = data["track"]
            let artist = data["artist"]
            if (track === "") {
                return
            }
            if (track === currentTrack && artist === currentArtist) {
                return
            }

            currentArtist = artist
            currentTrack = track
            $(".title").text(track)
            $(".subtitle").text(artist)
            $(".lyrics").text(`Loading lyrics for ${track} by ${artist}`)
            getLyrics()

        }, "json")

    }



    $("#scrobbleSwitch").change(function () {
        setScrobble(this.checked)
    })



    $(".dropdown .button").click(function (){
        let dropdown = $(this).parents('.dropdown');
        dropdown.toggleClass('is-active');

    });



    getPlayers()
    getUpdates()
    setInterval(getUpdates, 2 * 1000)
    setInterval(getPlayers, 15 * 1000)


});

