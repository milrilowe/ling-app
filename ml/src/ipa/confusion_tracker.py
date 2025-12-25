"""
Phoneme confusion matrix tracker.

Tracks which phonemes are substituted for others over multiple recordings.
"""

import json
from pathlib import Path
from typing import Dict, List, Tuple
from collections import defaultdict


class PhonemeConfusionTracker:
    """
    Tracks phoneme substitution patterns over time.

    Builds a confusion matrix showing:
    - How often each phoneme is said correctly
    - What phonemes are substituted when incorrect
    """

    def __init__(self, storage_path: str = None):
        """
        Initialize confusion tracker.

        Args:
            storage_path: Path to store confusion data (JSON file)
                         If None, uses in-memory only
        """
        self.storage_path = storage_path
        self.confusion_matrix: Dict[str, Dict[str, int]] = defaultdict(
            lambda: defaultdict(int)
        )

        # Load existing data if available
        if storage_path and Path(storage_path).exists():
            self.load()

    def record_alignment(
        self,
        alignment: List[Tuple[str, str, str]]
    ):
        """
        Record phoneme alignment results.

        Args:
            alignment: List of (expected, actual, type) tuples from aligner
        """
        for expected, actual, match_type in alignment:
            if match_type == 'delete':
                # Expected phoneme was not said
                # Record as substitution to '<missing>'
                self.confusion_matrix[expected]['<missing>'] += 1

            elif match_type == 'insert':
                # Extra phoneme was said
                # Record as insertion of the actual phoneme
                self.confusion_matrix['<extra>'][actual] += 1

            elif match_type == 'match':
                # Correct phoneme
                self.confusion_matrix[expected][expected] += 1

            elif match_type == 'substitute':
                # Phoneme substituted
                self.confusion_matrix[expected][actual] += 1

    def get_phoneme_stats(self, phoneme: str) -> Dict:
        """
        Get statistics for a specific phoneme.

        Args:
            phoneme: The phoneme to get stats for

        Returns:
            Dictionary with:
                - 'total': Total occurrences
                - 'correct': Number of correct pronunciations
                - 'accuracy': Percentage correct (0-100)
                - 'substitutions': Dict of {substitute: count}
        """
        if phoneme not in self.confusion_matrix:
            return {
                'total': 0,
                'correct': 0,
                'accuracy': 0.0,
                'substitutions': {}
            }

        phoneme_data = self.confusion_matrix[phoneme]
        total = sum(phoneme_data.values())
        correct = phoneme_data.get(phoneme, 0)
        accuracy = (correct / total * 100) if total > 0 else 0.0

        # Get substitutions (excluding correct)
        substitutions = {
            sub: count
            for sub, count in phoneme_data.items()
            if sub != phoneme
        }

        return {
            'total': total,
            'correct': correct,
            'accuracy': accuracy,
            'substitutions': substitutions
        }

    def get_all_stats(self) -> Dict[str, Dict]:
        """
        Get statistics for all phonemes.

        Returns:
            Dictionary mapping phoneme -> stats
        """
        all_stats = {}
        for phoneme in self.confusion_matrix.keys():
            all_stats[phoneme] = self.get_phoneme_stats(phoneme)

        return all_stats

    def get_worst_phonemes(self, n: int = 10) -> List[Tuple[str, float]]:
        """
        Get the N phonemes with worst accuracy.

        Args:
            n: Number of phonemes to return

        Returns:
            List of (phoneme, accuracy) tuples, sorted by accuracy (worst first)
        """
        phoneme_accuracies = []

        for phoneme in self.confusion_matrix.keys():
            if phoneme == '<extra>':
                continue  # Skip the special '<extra>' marker

            stats = self.get_phoneme_stats(phoneme)
            if stats['total'] >= 3:  # Only include if seen at least 3 times
                phoneme_accuracies.append((phoneme, stats['accuracy']))

        # Sort by accuracy (ascending = worst first)
        phoneme_accuracies.sort(key=lambda x: x[1])

        return phoneme_accuracies[:n]

    def get_summary(self) -> Dict:
        """
        Get overall summary statistics.

        Returns:
            Dictionary with:
                - 'total_phonemes': Number of unique phonemes
                - 'total_samples': Total phoneme comparisons
                - 'overall_accuracy': Overall accuracy percentage
                - 'worst_phonemes': Top 5 worst phonemes
        """
        total_samples = 0
        total_correct = 0

        for phoneme in self.confusion_matrix.keys():
            if phoneme == '<extra>':
                continue

            stats = self.get_phoneme_stats(phoneme)
            total_samples += stats['total']
            total_correct += stats['correct']

        overall_accuracy = (total_correct / total_samples * 100) if total_samples > 0 else 0.0

        return {
            'total_phonemes': len(self.confusion_matrix),
            'total_samples': total_samples,
            'overall_accuracy': overall_accuracy,
            'worst_phonemes': self.get_worst_phonemes(5)
        }

    def save(self):
        """Save confusion matrix to file."""
        if not self.storage_path:
            return

        # Convert defaultdict to regular dict for JSON serialization
        data = {
            phoneme: dict(substitutions)
            for phoneme, substitutions in self.confusion_matrix.items()
        }

        Path(self.storage_path).parent.mkdir(parents=True, exist_ok=True)

        with open(self.storage_path, 'w', encoding='utf-8') as f:
            json.dump(data, f, ensure_ascii=False, indent=2)

    def load(self):
        """Load confusion matrix from file."""
        if not self.storage_path or not Path(self.storage_path).exists():
            return

        with open(self.storage_path, 'r', encoding='utf-8') as f:
            data = json.load(f)

        # Convert back to defaultdict
        self.confusion_matrix = defaultdict(lambda: defaultdict(int))
        for phoneme, substitutions in data.items():
            for sub, count in substitutions.items():
                self.confusion_matrix[phoneme][sub] = count

    def clear(self):
        """Clear all confusion data."""
        self.confusion_matrix = defaultdict(lambda: defaultdict(int))
        if self.storage_path and Path(self.storage_path).exists():
            Path(self.storage_path).unlink()
