import sys
import yt_dlp


def get_stream_url(url: str) -> str:
    # Configure yt-dlp to get best quality HLS stream
    ydl_opts = {
        'format': 'best[ext=m3u8]/best',  # Prefer HLS, fallback to best
        'quiet': True,
        'no_warnings': True,
    }
    
    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        result = ydl.extract_info(url, download=False)
        
        # Get the selected format URL
        if 'url' in result:
            return result['url']
        
        # Fallback: find best HLS format manually
        formats = result.get('formats', [])
        
        # Filter for HLS formats and sort by height (resolution)
        hls_formats = [
            f for f in formats 
            if f.get('protocol') in ['m3u8', 'm3u8_native'] and f.get('height')
        ]
        
        if hls_formats:
            # Get highest resolution HLS format
            best_format = max(hls_formats, key=lambda x: x.get('height', 0))
            return best_format['url']
        
        # Last resort: return first available format
        return formats[0]['url'] if formats else ""

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python yt_dlp.py <youtube_url>", file=sys.stderr)
        sys.exit(1)
    
    yt_url = sys.argv[1]
    print(get_stream_url(yt_url))