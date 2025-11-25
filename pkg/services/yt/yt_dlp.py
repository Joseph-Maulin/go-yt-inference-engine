import sys
import yt_dlp


def get_stream_url(url: str) -> str:
    with yt_dlp.YoutubeDL({}) as ydl:
        result = ydl.extract_info(url, download=False)
        return result['formats'][0]['url']

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python yt_dlp.py <youtube_url>", file=sys.stderr)
        sys.exit(1)
    
    yt_url = sys.argv[1]
    print(get_stream_url(yt_url))