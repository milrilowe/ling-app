"""
TextGrid parser for MFA output.

Parses Praat TextGrid files and converts to JSON format
with dual ARPAbet/IPA phoneme representation.
"""

import logging
from typing import Any

import textgrid

from ..ipa.arpabet_ipa import arpabet_to_ipa

logger = logging.getLogger(__name__)


def parse_textgrid(path: str) -> dict[str, Any]:
    """
    Parse MFA TextGrid output into JSON structure.

    MFA outputs a TextGrid with two tiers:
    - "words": Word-level intervals
    - "phones": Phone-level intervals (in ARPAbet format)

    Args:
        path: Path to TextGrid file

    Returns:
        Dict with:
        - words: List of {word, start, end}
        - phones: List of {arpabet, ipa, start, end}
        - duration: Total audio duration
    """
    logger.info(f"Parsing TextGrid: {path}")

    tg = textgrid.TextGrid.fromFile(path)

    words: list[dict] = []
    phones: list[dict] = []
    duration = float(tg.maxTime)

    for tier in tg:
        if tier.name == "words":
            for interval in tier:
                # Skip empty intervals (silence)
                if interval.mark and interval.mark.strip():
                    words.append({
                        "word": interval.mark,
                        "start": round(float(interval.minTime), 3),
                        "end": round(float(interval.maxTime), 3),
                    })

        elif tier.name == "phones":
            for interval in tier:
                # Skip empty intervals and silence markers
                mark = interval.mark
                if mark and mark.strip() and mark not in ("", "sil", "sp", "spn"):
                    arpabet = mark  # e.g., "HH", "AH0", "L"
                    phones.append({
                        "arpabet": arpabet,
                        "ipa": arpabet_to_ipa(arpabet),
                        "start": round(float(interval.minTime), 3),
                        "end": round(float(interval.maxTime), 3),
                    })

    logger.info(
        f"Parsed TextGrid: {len(words)} words, {len(phones)} phones, "
        f"duration: {duration:.2f}s"
    )

    return {
        "words": words,
        "phones": phones,
        "duration": round(duration, 3),
    }
