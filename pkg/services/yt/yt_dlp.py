import yt_dlp


def get_stream_url(url: str) -> str:
    with yt_dlp.YoutubeDL({}) as ydl:
        result = ydl.extract_info(url, download=False)
        return result['formats'][0]['url']

if __name__ == "__main__":
    yt_url = "https://www.youtube.com/watch?v=XDThHUawq6E"
    print(get_stream_url(yt_url))