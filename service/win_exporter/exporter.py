import asyncio
import json
from winrt.windows.media.control import GlobalSystemMediaTransportControlsSessionManager as MediaManager
# https://stackoverflow.com/q/65011660/


async def get_media_info():
    sessions = await MediaManager.request_async()

    # This source_app_user_model_id check and if statement is optional
    # Use it if you want to only get a certain player/program's media
    # (e.g. only chrome.exe's media not any other program's).

    # To get the ID, use a breakpoint() to run sessions.get_current_session()
    # while the media you want to get is playing.
    # Then set TARGET_ID to the string this call returns.

    current_session = sessions.get_current_session()
    #set_trace()
    if current_session:  # there needs to be a media session running

            info = await current_session.try_get_media_properties_async()

            # song_attr[0] != '_' ignores system attributes
            info_dict = {song_attr: info.__getattribute__(song_attr) for song_attr in dir(info) if song_attr[0] != '_'}

            # converts winrt vector to list
            info_dict['genres'] = list(info_dict['genres'])

            return info_dict
    else:
        return {}



if __name__ == '__main__':
    current_media_info = asyncio.run(get_media_info())
    n = {
        "artist": current_media_info.get("artist"),
        "title": current_media_info.get("title")
    }
    print(json.dumps(n))
