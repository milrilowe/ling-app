#!/usr/bin/env python3
"""
Generate audio files for phoneme example words using ElevenLabs TTS.
Run once to create static audio files for the pronunciation dashboard.

Usage:
    ELEVENLABS_API_KEY=your_key python scripts/generate_phoneme_audio.py
"""

import os
import requests
import time
from pathlib import Path

# ElevenLabs API configuration
API_KEY = os.environ.get("ELEVENLABS_API_KEY")
VOICE_ID = "21m00Tcm4TlvDq8ikWAM"  # Rachel voice
BASE_URL = "https://api.elevenlabs.io/v1"

# Output directory
OUTPUT_DIR = Path(__file__).parent.parent / "web" / "public" / "audio" / "phonemes"

# All 39 phoneme example words (must match phonemes.ts)
PHONEME_EXAMPLES = [
    # Vowels
    ("ɑ", "father"),
    ("æ", "cat"),
    ("ʌ", "but"),
    ("ɔ", "thought"),
    ("ɛ", "bed"),
    ("ɝ", "bird"),
    ("ɪ", "bit"),
    ("i", "bee"),
    ("ʊ", "book"),
    ("u", "blue"),
    # Diphthongs
    ("aʊ", "cow"),
    ("aɪ", "buy"),
    ("eɪ", "say"),
    ("oʊ", "go"),
    ("ɔɪ", "boy"),
    # Stops
    ("b", "boy"),
    ("d", "dog"),
    ("g", "go"),
    ("k", "cat"),
    ("p", "pot"),
    ("t", "top"),
    # Fricatives
    ("ð", "this"),
    ("f", "fish"),
    ("s", "sun"),
    ("ʃ", "she"),
    ("θ", "think"),
    ("v", "voice"),
    ("z", "zoo"),
    ("ʒ", "measure"),
    ("h", "hello"),
    # Affricates
    ("tʃ", "church"),
    ("dʒ", "judge"),
    # Nasals
    ("m", "man"),
    ("n", "no"),
    ("ŋ", "sing"),
    # Liquids
    ("l", "love"),
    ("ɹ", "red"),
    # Glides
    ("w", "water"),
    ("j", "yes"),
]


def generate_audio(text: str, output_path: Path) -> bool:
    """Generate audio for a word using ElevenLabs TTS."""
    url = f"{BASE_URL}/text-to-speech/{VOICE_ID}"

    headers = {
        "Accept": "audio/mpeg",
        "Content-Type": "application/json",
        "xi-api-key": API_KEY,
    }

    data = {
        "text": text,
        "model_id": "eleven_multilingual_v2",
        "voice_settings": {
            "stability": 0.5,
            "similarity_boost": 0.75,
        },
    }

    response = requests.post(url, json=data, headers=headers)

    if response.status_code == 200:
        output_path.write_bytes(response.content)
        return True
    else:
        print(f"  Error: {response.status_code} - {response.text}")
        return False


def main():
    if not API_KEY:
        print("Error: ELEVENLABS_API_KEY environment variable not set")
        print("Usage: ELEVENLABS_API_KEY=your_key python scripts/generate_phoneme_audio.py")
        return

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    # Track unique words to avoid duplicates (e.g., "boy" appears twice)
    generated_words = set()

    print(f"Generating audio for {len(PHONEME_EXAMPLES)} phonemes...")
    print(f"Output directory: {OUTPUT_DIR}")
    print()

    for ipa, word in PHONEME_EXAMPLES:
        if word in generated_words:
            print(f"  /{ipa}/ -> {word}.mp3 (already generated)")
            continue

        output_path = OUTPUT_DIR / f"{word}.mp3"

        if output_path.exists():
            print(f"  /{ipa}/ -> {word}.mp3 (exists, skipping)")
            generated_words.add(word)
            continue

        print(f"  /{ipa}/ -> {word}.mp3 ... ", end="", flush=True)

        if generate_audio(word, output_path):
            print("done")
            generated_words.add(word)
        else:
            print("FAILED")

        # Rate limiting - ElevenLabs has limits
        time.sleep(0.5)

    print()
    print(f"Generated {len(generated_words)} unique audio files")


if __name__ == "__main__":
    main()
