let currentTrack = ""
let currentArtist = ""
let currentPlayers = []


let RATE_LIMIT_IN_MS = 1000;
let NUMBER_OF_REQUESTS_ALLOWED = 1;
let NUMBER_OF_REQUESTS = 0;
let coverArtQueue = []


function getSongHTML(track, artist, albumArt, mbid, artistMbid) {

    let albumArtUrl = "http://bulma.io/images/placeholders/96x96.png"
    if (albumArt !== "") {
        albumArtUrl = albumArt
    }


    return `<div class="card mb-3">
        <div class="card-content">
            <div class="media">
                <div class="media-left">
                    <figure class="image is-48x48">
                        <img 
                        data-artist="${artist}"
                        data-track="${track}"
                        src="${albumArtUrl}" alt="Album art of ${track} by ${artist}">
                    </figure>
                </div>
                <div class="media-content">
                    <p class="title is-4">${track}</p>
                    <p class="subtitle is-6">${artist}</p>
                </div>
            </div>
            
        </div>
        <footer class="card-footer">
            <a class="card-footer-item" data-artist="${artist}" data-track="${track}">
              <span class="icon"><i class="fas fa-play"></i></span>Play Next</a>
            <a class="card-footer-item" data-artist="${artist}" data-track="${track}">
              <span class="icon"><i class="fas fa-download"></i></span>Download</a>
            <a class="card-footer-item" data-artist="${artist}" data-track="${track}">
              <span class="icon"><i class="fas fa-heart"></i></span>Like</a>
        </footer>
    </div>`
}

$(document).ready(function() {

    // Check for click events on the navbar burger icon
    $(".navbar-burger").click(function() {

        // Toggle the "is-active" class on both the "navbar-burger" and the "navbar-menu"
        $(".navbar-burger").toggleClass("is-active");
        $(".navbar-menu").toggleClass("is-active");
    });


    $.get("/api/v1/user/logged-in", function(data) {
        console.log("Received data from /user/logged-in")
        if (data["logged_in"] === false) {
            console.log("User is not logged in, disabling scrobble switch")
            window.location.replace("/login")
        } else {
            $("#navBarLoginButton").remove()
        }
    }, "json")

    function getSimilarSongs() {
        console.log("Trying to fetch Lyrics")
        $.get("/api/v1/song/similar", function (data) {
            console.log("Received similar songs")
            let similarContainer = $("#similar-songs-container")
            similarContainer.empty()
            data.forEach(function (item) {
                if (item["track"] === "") {
                    return
                }
                similarContainer.append(getSongHTML(item["track"], item["artist"], item["album_art"], item["mbid"]))
                console.log(item["track"], item["artist"], item["album_art"], item["mbid"])

            })
            data.forEach(function(item) {
                let track = item["track"]
                let artistMbid = item["artist_mbid"]

                if (artistMbid !== "" && artistMbid !== undefined) {

                    coverArtQueue.push(
                        function () {
                            const mbrzUrl = `http://musicbrainz.org/ws/2/release?query=${encodeURI(track)}%20AND%20arid:${artistMbid}&fmt=json`
                            console.log(mbrzUrl)

                            $.get(mbrzUrl, function (data) {
                                if (data["releases"] === undefined || data["releases"].length === 0) {
                                    return
                                }
                                const mbId = data["releases"][0]["id"]
                                $.get(`http://coverartarchive.org/release/${mbId}`, function (data) {
                                    let albumArtImage = data["images"][0]["image"]
                                    console.log(albumArtImage)
                                    $(`[data-track="${track}"]`).attr("src", albumArtImage)
                                }, "json")

                            }, "json")
                        }
                    )

                }
            })
            // $(".lyrics").html(data.replaceAll("\n", "<br>"))
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
            coverArtQueue = []

            $(".subtitle").text(`similar to your ${track} by ${artist}`)
            getSimilarSongs()

        }, "json")

    }


    $("#scrobbleSwitch").change(function () {
        setScrobble(this.checked)
    })



    $(".dropdown .button").click(function (){
        let dropdown = $(this).parents('.dropdown');
        dropdown.toggleClass('is-active');

    });

    getUpdates()
    setInterval(getUpdates, 2 * 1000)
});


function GetCovertArtRateLimited() {
    if (coverArtQueue.length === 0) {
        return
    }
    let fx = coverArtQueue.shift()

    fx()
}


setInterval(function()
{
    GetCovertArtRateLimited()

}, RATE_LIMIT_IN_MS);

