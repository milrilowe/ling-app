#!/usr/bin/env python3
"""
Download isolated IPA phoneme sounds from jbdowse.com/ipa.

This script:
1. Fetches the HTML from jbdowse.com/ipa
2. Parses to extract IPA symbol to file ID mappings
3. Downloads isolated sounds (WAV format)
4. Converts to MP3 for smaller file size
5. Saves with URL-safe filenames

Usage:
    python scripts/download_ipa_audio.py

Requirements:
    pip install requests beautifulsoup4 pydub

Note: Requires ffmpeg installed for WAV to MP3 conversion.
"""

import os
import re
import time
import json
import unicodedata
from pathlib import Path
from typing import Optional

import requests
from bs4 import BeautifulSoup

try:
    from pydub import AudioSegment
    PYDUB_AVAILABLE = True
except ImportError:
    PYDUB_AVAILABLE = False
    print("Warning: pydub not available, will keep WAV format")


# Output directory
OUTPUT_DIR = Path(__file__).parent.parent / "web" / "public" / "audio" / "ipa"

# Source URL
SOURCE_URL = "https://jbdowse.com/ipa/"
AUDIO_BASE_URL = "https://jbdowse.com/ipa/s/"

# Rate limiting
REQUEST_DELAY = 0.2  # seconds between requests


def sanitize_filename(ipa: str) -> str:
    """
    Convert IPA symbol to a safe filename.

    IPA symbols often contain Unicode characters that are problematic
    for filenames. This function creates URL-safe alternatives.
    """
    # Simple ASCII characters can be used directly
    if ipa.isascii() and ipa.isalnum():
        return ipa

    # Map common IPA symbols to readable names
    IPA_NAMES = {
        # Vowels
        'ɑ': 'open_back_unrounded',
        'æ': 'near_open_front_unrounded',
        'ɐ': 'near_open_central',
        'ɒ': 'open_back_rounded',
        'ə': 'schwa',
        'ɚ': 'r_colored_schwa',
        'ɵ': 'close_mid_central_rounded',
        'ɘ': 'close_mid_central_unrounded',
        'ɛ': 'open_mid_front_unrounded',
        'ɜ': 'open_mid_central_unrounded',
        'ɝ': 'r_colored_open_mid_central',
        'ɞ': 'open_mid_central_rounded',
        'ɤ': 'close_mid_back_unrounded',
        'ɪ': 'near_close_front_unrounded',
        'ɨ': 'close_central_unrounded',
        'ɔ': 'open_mid_back_rounded',
        'ʊ': 'near_close_back_rounded',
        'ʉ': 'close_central_rounded',
        'ʌ': 'open_mid_back_unrounded',
        'ʏ': 'near_close_front_rounded',
        'œ': 'open_mid_front_rounded',
        'ø': 'close_mid_front_rounded',

        # Consonants
        'ʔ': 'glottal_stop',
        'ʕ': 'pharyngeal_fricative_voiced',
        'ħ': 'pharyngeal_fricative_voiceless',
        'ʜ': 'epiglottal_fricative_voiceless',
        'ʢ': 'epiglottal_fricative_voiced',
        'ʡ': 'epiglottal_stop',
        'θ': 'dental_fricative_voiceless',
        'ð': 'dental_fricative_voiced',
        'ʃ': 'postalveolar_fricative_voiceless',
        'ʒ': 'postalveolar_fricative_voiced',
        'ʂ': 'retroflex_fricative_voiceless',
        'ʐ': 'retroflex_fricative_voiced',
        'ɕ': 'alveolo_palatal_fricative_voiceless',
        'ʑ': 'alveolo_palatal_fricative_voiced',
        'ç': 'palatal_fricative_voiceless',
        'ʝ': 'palatal_fricative_voiced',
        'χ': 'uvular_fricative_voiceless',
        'ʁ': 'uvular_fricative_voiced',
        'ɣ': 'velar_fricative_voiced',
        'ɸ': 'bilabial_fricative_voiceless',
        'β': 'bilabial_fricative_voiced',
        'ɬ': 'lateral_fricative_voiceless',
        'ɮ': 'lateral_fricative_voiced',
        'ŋ': 'velar_nasal',
        'ɲ': 'palatal_nasal',
        'ɳ': 'retroflex_nasal',
        'ɴ': 'uvular_nasal',
        'ɱ': 'labiodental_nasal',
        'ʙ': 'bilabial_trill',
        'ʀ': 'uvular_trill',
        'ɾ': 'alveolar_tap',
        'ɽ': 'retroflex_flap',
        'ɹ': 'alveolar_approximant',
        'ɻ': 'retroflex_approximant',
        'ʋ': 'labiodental_approximant',
        'ɰ': 'velar_approximant',
        'ɭ': 'retroflex_lateral',
        'ʎ': 'palatal_lateral',
        'ʟ': 'velar_lateral',
        'ɖ': 'retroflex_stop_voiced',
        'ʈ': 'retroflex_stop_voiceless',
        'ɟ': 'palatal_stop_voiced',
        'ɡ': 'velar_stop_voiced',
        'ɢ': 'uvular_stop_voiced',
        'ʛ': 'uvular_implosive_voiced',
        'ɓ': 'bilabial_implosive',
        'ɗ': 'alveolar_implosive',
        'ʄ': 'palatal_implosive',
        'ɠ': 'velar_implosive',
        'ʘ': 'bilabial_click',
        'ǀ': 'dental_click',
        'ǃ': 'postalveolar_click',
        'ǂ': 'palatoalveolar_click',
        'ǁ': 'lateral_click',

        # Affricates (common combinations)
        'tʃ': 'voiceless_postalveolar_affricate',
        'dʒ': 'voiced_postalveolar_affricate',
        'ts': 'voiceless_alveolar_affricate',
        'dz': 'voiced_alveolar_affricate',

        # Diacritics and modifiers
        'ˈ': 'primary_stress',
        'ˌ': 'secondary_stress',
        'ː': 'long',
        'ˑ': 'half_long',
        '̃': 'nasalized',
        'ʰ': 'aspirated',
        'ʷ': 'labialized',
        'ʲ': 'palatalized',
        'ˠ': 'velarized',
        'ˤ': 'pharyngealized',
    }

    if ipa in IPA_NAMES:
        return IPA_NAMES[ipa]

    # For compound symbols or unknown ones, use Unicode code points
    # This ensures uniqueness and reversibility
    codes = '_'.join(f'u{ord(c):04x}' for c in ipa)
    return codes


def fetch_html() -> str:
    """Fetch the IPA chart HTML."""
    print(f"Fetching {SOURCE_URL}...")
    response = requests.get(SOURCE_URL, timeout=30)
    response.raise_for_status()
    return response.text


def parse_phoneme_mappings(html: str) -> dict[str, tuple[str, str]]:
    """
    Parse HTML to extract phoneme mappings.

    Returns:
        Dict mapping file_id to (ipa_symbol, description)
        We only want the first button of each group (isolated sound).
    """
    soup = BeautifulSoup(html, 'html.parser')
    mappings = {}

    # Find all table cells with phoneme data
    # Structure: <td> contains <p>IPA symbol</p> and <button data-src="XXX">
    for td in soup.find_all('td'):
        p_tag = td.find('p')
        if not p_tag:
            continue

        ipa_symbol = p_tag.get_text(strip=True)
        if not ipa_symbol:
            continue

        # Get all buttons in this cell
        buttons = td.find_all('button', attrs={'data-src': True})
        if not buttons:
            continue

        # First button is the isolated sound
        first_button = buttons[0]
        file_id = first_button.get('data-src')

        if file_id:
            # Get parent row's first cell for description (manner of articulation)
            row = td.find_parent('tr')
            if row:
                first_cell = row.find('td')
                description = first_cell.get_text(strip=True) if first_cell else ''
            else:
                description = ''

            mappings[file_id] = (ipa_symbol, description)

    return mappings


def download_audio(file_id: str, output_path: Path) -> bool:
    """Download a single audio file."""
    url = f"{AUDIO_BASE_URL}{file_id}.wav"

    try:
        response = requests.get(url, timeout=30)
        response.raise_for_status()

        # Save as WAV first
        wav_path = output_path.with_suffix('.wav')
        wav_path.write_bytes(response.content)

        # Convert to MP3 if pydub is available
        if PYDUB_AVAILABLE:
            try:
                audio = AudioSegment.from_wav(str(wav_path))
                audio.export(str(output_path), format='mp3', bitrate='128k')
                wav_path.unlink()  # Remove WAV after conversion
                return True
            except Exception as e:
                print(f"    MP3 conversion failed: {e}, keeping WAV")
                wav_path.rename(output_path.with_suffix('.wav'))
                return True
        else:
            wav_path.rename(output_path.with_suffix('.wav'))
            return True

    except requests.RequestException as e:
        print(f"    Download failed: {e}")
        return False


def main():
    # Create output directory
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    # Fetch and parse HTML
    html = fetch_html()
    mappings = parse_phoneme_mappings(html)

    print(f"Found {len(mappings)} phonemes to download")
    print(f"Output directory: {OUTPUT_DIR}")
    print()

    # Track what we download for the TypeScript mapping file
    downloaded = {}
    failed = []
    skipped = []

    for i, (file_id, (ipa_symbol, description)) in enumerate(mappings.items(), 1):
        safe_name = sanitize_filename(ipa_symbol)
        ext = '.mp3' if PYDUB_AVAILABLE else '.wav'
        output_path = OUTPUT_DIR / f"{safe_name}{ext}"

        # Skip if already exists
        if output_path.exists():
            print(f"[{i}/{len(mappings)}] /{ipa_symbol}/ -> {safe_name}{ext} (exists, skipping)")
            downloaded[ipa_symbol] = f"{safe_name}{ext}"
            skipped.append(ipa_symbol)
            continue

        print(f"[{i}/{len(mappings)}] /{ipa_symbol}/ -> {safe_name}{ext} ... ", end='', flush=True)

        if download_audio(file_id, output_path):
            print("done")
            downloaded[ipa_symbol] = f"{safe_name}{ext}"
        else:
            print("FAILED")
            failed.append(ipa_symbol)

        # Rate limiting
        time.sleep(REQUEST_DELAY)

    print()
    print(f"Downloaded: {len(downloaded) - len(skipped)}")
    print(f"Skipped (existing): {len(skipped)}")
    print(f"Failed: {len(failed)}")

    if failed:
        print(f"Failed phonemes: {failed}")

    # Save mapping as JSON for reference
    mapping_path = OUTPUT_DIR / "_mapping.json"
    with open(mapping_path, 'w', encoding='utf-8') as f:
        json.dump({
            'source': SOURCE_URL,
            'generated': time.strftime('%Y-%m-%d %H:%M:%S'),
            'count': len(downloaded),
            'mappings': downloaded
        }, f, indent=2, ensure_ascii=False)

    print(f"\nMapping saved to: {mapping_path}")

    # Generate TypeScript file
    ts_path = Path(__file__).parent.parent / "web" / "src" / "data" / "ipa-sounds.ts"
    generate_typescript(downloaded, ts_path)
    print(f"TypeScript file saved to: {ts_path}")


def generate_typescript(mappings: dict[str, str], output_path: Path):
    """Generate TypeScript mapping file."""
    lines = [
        '// Auto-generated by scripts/download_ipa_audio.py',
        '// Maps IPA symbols to audio filenames in /audio/ipa/',
        '',
        'export const IPA_AUDIO: Record<string, string> = {',
    ]

    for ipa, filename in sorted(mappings.items()):
        # Escape any special characters in the IPA symbol for the key
        escaped_ipa = ipa.replace('\\', '\\\\').replace("'", "\\'")
        lines.append(f"  '{escaped_ipa}': '{filename}',")

    lines.extend([
        '};',
        '',
        '/**',
        ' * Get the URL for an IPA sound file.',
        ' * Returns null if no audio exists for the given IPA symbol.',
        ' */',
        'export function getIpaSoundUrl(ipa: string): string | null {',
        '  const file = IPA_AUDIO[ipa];',
        '  return file ? `/audio/ipa/${file}` : null;',
        '}',
        '',
    ])

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text('\n'.join(lines), encoding='utf-8')


if __name__ == "__main__":
    main()
