"""
Text to IPA conversion using gruut phonemizer.

Uses gruut to convert text to IPA, matching the format used to train
the Whisper IPA model.
"""

from typing import List, Optional

import gruut


class GruutIPAConverter:
    """
    Converts text to IPA phonetic transcription using gruut.

    gruut is the same tool used to generate training data for the
    neuralang/ipa-whisper-small model, ensuring format compatibility.
    """

    def __init__(self, language: str = "en-us"):
        """
        Initialize Gruut IPA converter.

        Args:
            language: Language code (e.g., 'en-us', 'en-gb', 'de', 'es')
                     Note: Language-specific gruut packages must be installed
                     (e.g., gruut[en] for English)
        """
        self.language = language

    def text_to_ipa(self, text: str) -> str:
        """
        Convert text to IPA transcription.

        Args:
            text: Input text to convert (e.g., "Hello world")

        Returns:
            IPA transcription string (e.g., "hɛloʊ wɜrld")

        Raises:
            ValueError: If language is not supported or not installed
        """
        if not text or not text.strip():
            return ""

        try:
            # Process text through gruut
            phonemes = []

            for sentence in gruut.sentences(text, lang=self.language):
                for word in sentence:
                    # word.phonemes contains IPA phonemes for this word
                    if word.phonemes:
                        phonemes.extend(word.phonemes)
                        # Add space between words
                        phonemes.append(" ")

            # Join phonemes into string and clean up extra spaces
            ipa_string = "".join(phonemes).strip()

            # Clean up multiple spaces
            while "  " in ipa_string:
                ipa_string = ipa_string.replace("  ", " ")

            return ipa_string

        except Exception as e:
            raise ValueError(
                f"Failed to convert text to IPA for language '{self.language}'. "
                f"Is gruut[{self.language.split('-')[0]}] installed? "
                f"Error: {str(e)}"
            ) from e

    def text_to_phoneme_list(self, text: str) -> List[str]:
        """
        Convert text to list of individual phonemes.

        Args:
            text: Input text to convert

        Returns:
            List of IPA phonemes
        """
        phonemes = []

        for sentence in gruut.sentences(text, lang=self.language):
            for word in sentence:
                if word.phonemes:
                    phonemes.extend(word.phonemes)

        return phonemes


def text_to_ipa(text: str, language: str = "en-us") -> str:
    """
    Convenience function to convert text to IPA.

    Args:
        text: Input text
        language: Language code (default: en-us)

    Returns:
        IPA transcription string
    """
    converter = GruutIPAConverter(language=language)
    return converter.text_to_ipa(text)
