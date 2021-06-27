let currentTrack = ""
let currentArtist = ""
let currentPlayers = []


function playerChangedCallback(item) {
    console.log("Requesting server to change player")
    let data = {
        "dbus_id": item.text
    }
    $.postJSON("/player", data, function () {
        console.log(`Player changed successfully, received ${data} as player name`)
    })
    $(".dropdown").removeClass("is-active")

}

function setScrobble(enabled) {
    console.log("Requesting server to change scrobble settings")
    $.post(`/prefs/scrobble/${enabled}`, "", function () {
        console.log("Player changed successfully")
    }, "text")
}


function getLyrics() {
    console.log("Trying to fetch Lyrics")
    $.get("/updates/lyrics", function (data) {
        console.log("Received lyrics")
        $(".lyrics").html(data.replaceAll("\n", "<br>"))
    }, "text")

}

function getPlayers() {
    $.get("/updates/players", function (data) {
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

    $.get("/updates/song", function (data) {
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
    dropdown.focusout(function() {
        $(this).removeClass('is-active');
    });
});



getPlayers()
getUpdates()
setInterval(getUpdates, 2 * 1000)
setInterval(getPlayers, 15 * 1000)