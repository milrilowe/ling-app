#!/usr/bin/env python3
"""
Speech Service CLI

Simple CLI to test audio-to-IPA and text-to-IPA conversion.
"""

import sys
from pathlib import Path

import click
from rich.console import Console
from rich.table import Table
from rich.panel import Panel

from src.audio.loader import AudioLoader
from src.ipa.audio_to_ipa import WhisperIPAConverter
from src.ipa.text_to_ipa import GruutIPAConverter
from src.ipa.normalizer import IPANormalizer
from src.ipa.aligner import PhonemeAligner


console = Console()


@click.command()
@click.option(
    '--audio',
    required=True,
    type=click.Path(exists=True),
    help='Path to audio file (WebM, WAV, MP3, etc.)'
)
@click.option(
    '--text',
    required=True,
    type=str,
    help='Text transcript of the audio'
)
@click.option(
    '--language',
    default='en-us',
    type=str,
    help='Language code for text-to-IPA (default: en-us)'
)
@click.option(
    '--device',
    default=None,
    type=str,
    help='Device for Whisper model (cpu/cuda, default: auto-detect)'
)
def compare_pronunciation(audio: str, text: str, language: str, device: str):
    """
    Compare pronunciation between audio and text transcript.

    Converts both to IPA and displays the results.

    \b
    Example:
        python cli.py --audio sample.webm --text "Hello world"
    """
    console.print("\n[bold cyan]Speech Service - IPA Comparison[/bold cyan]\n")

    try:
        # 1. Load audio
        console.print("[yellow]📁 Loading audio file...[/yellow]")
        loader = AudioLoader()
        audio_array, sample_rate = loader.load_audio(audio)
        console.print(f"[green]✓ Loaded audio: {sample_rate}Hz, {len(audio_array)} samples[/green]")

        # 2. Convert audio to IPA
        console.print("\n[yellow]🎤 Converting audio to IPA with Whisper...[/yellow]")
        whisper_converter = WhisperIPAConverter(device=device)
        audio_ipa = whisper_converter.audio_to_ipa(audio_array, sample_rate)
        console.print(f"[green]✓ Audio IPA transcription complete[/green]")

        # 3. Convert text to IPA
        console.print("\n[yellow]📝 Converting text to IPA with gruut...[/yellow]")
        gruut_converter = GruutIPAConverter(language=language)
        text_ipa = gruut_converter.text_to_ipa(text)
        console.print(f"[green]✓ Text IPA conversion complete[/green]")

        # 4. Align phonemes
        console.print("\n[yellow]🔍 Aligning phonemes...[/yellow]")
        aligner = PhonemeAligner()
        alignment = aligner.align(audio_ipa, text_ipa)
        console.print(f"[green]✓ Phoneme alignment complete[/green]")

        # 5. Normalize and compare
        normalizer = IPANormalizer()
        comparison = normalizer.compare(audio_ipa, text_ipa, ignore_whitespace=True)

        # 6. Display results
        console.print("\n[bold cyan]Results:[/bold cyan]\n")

        # Create comparison table
        table = Table(title="IPA Comparison", show_header=True, header_style="bold magenta")
        table.add_column("Source", style="cyan", width=20)
        table.add_column("IPA Output", style="green")

        table.add_row("Original Text", text)
        table.add_row("Audio (Whisper)", audio_ipa)
        table.add_row("Text (gruut)", text_ipa)
        table.add_row("", "")  # Empty row
        table.add_row("Normalized Audio", comparison['normalized_1'])
        table.add_row("Normalized Text", comparison['normalized_2'])

        console.print(table)

        # Match statistics
        stats_text = f"""
[bold]Exact Match:[/bold] {comparison['exact_match']}
[bold]Audio IPA Length:[/bold] {comparison['length_1']} characters
[bold]Text IPA Length:[/bold] {comparison['length_2']} characters
[bold]Differences:[/bold] {len(comparison['differences'])} character(s)
"""
        console.print(Panel(stats_text, title="Statistics", border_style="blue"))

        # Show differences if any
        if not comparison['exact_match'] and comparison['differences']:
            console.print("\n[bold red]Character-level differences:[/bold red]")
            diff_table = Table(show_header=True, header_style="bold red")
            diff_table.add_column("Position", style="yellow", width=10)
            diff_table.add_column("Audio", style="cyan", width=15)
            diff_table.add_column("Text", style="green", width=15)

            # Show first 10 differences
            for i, (pos, char1, char2) in enumerate(comparison['differences'][:10]):
                char1_display = repr(char1) if char1 else "[missing]"
                char2_display = repr(char2) if char2 else "[missing]"
                diff_table.add_row(str(pos), char1_display, char2_display)

            if len(comparison['differences']) > 10:
                diff_table.add_row("...", "...", "...")

            console.print(diff_table)

        # Visual alignment
        if not comparison['exact_match']:
            console.print("\n[bold cyan]Visual Alignment:[/bold cyan]")
            aligned1, match_str, aligned2 = normalizer.get_alignment(audio_ipa, text_ipa)

            # Show first 80 characters
            max_display = 80
            if len(aligned1) > max_display:
                aligned1 = aligned1[:max_display] + "..."
                aligned2 = aligned2[:max_display] + "..."
                match_str = match_str[:max_display] + "..."

            console.print(f"Audio: {aligned1}")
            console.print(f"       {match_str}")
            console.print(f"Text:  {aligned2}")

        # Phoneme-level analysis
        console.print("\n[bold cyan]Phoneme Analysis:[/bold cyan]")

        # Count phoneme types
        matches = sum(1 for _, _, t in alignment if t == 'match')
        substitutions = sum(1 for _, _, t in alignment if t == 'substitute')
        deletions = sum(1 for _, _, t in alignment if t == 'delete')
        insertions = sum(1 for _, _, t in alignment if t == 'insert')
        total = len(alignment)

        phoneme_stats_text = f"""
[bold]Total Phonemes:[/bold] {total}
[bold]Matches:[/bold] {matches} ({matches/total*100:.1f}% if total > 0 else 0)
[bold]Substitutions:[/bold] {substitutions}
[bold]Deletions (not said):[/bold] {deletions}
[bold]Insertions (extra):[/bold] {insertions}
"""
        console.print(Panel(phoneme_stats_text, title="Phoneme Statistics", border_style="green"))

        # Show substitution details
        if substitutions > 0:
            console.print("\n[bold yellow]Phoneme Substitutions:[/bold yellow]")
            sub_table = Table(show_header=True, header_style="bold yellow")
            sub_table.add_column("Expected", style="cyan")
            sub_table.add_column("Actually Said", style="red")

            for expected, actual, match_type in alignment:
                if match_type == 'substitute':
                    sub_table.add_row(expected, actual)

            console.print(sub_table)

        console.print("\n[bold green]✓ Comparison complete![/bold green]\n")

    except FileNotFoundError as e:
        console.print(f"\n[bold red]Error:[/bold red] {str(e)}\n", style="red")
        sys.exit(1)
    except ValueError as e:
        console.print(f"\n[bold red]Error:[/bold red] {str(e)}\n", style="red")
        console.print(
            "[yellow]Tip: Make sure gruut[en] is installed for English support.[/yellow]\n"
        )
        sys.exit(1)
    except Exception as e:
        console.print(f"\n[bold red]Unexpected error:[/bold red] {str(e)}\n", style="red")
        import traceback
        console.print(traceback.format_exc())
        sys.exit(1)


@click.group()
def cli():
    """Speech Service CLI tools."""
    pass


cli.add_command(compare_pronunciation)


if __name__ == '__main__':
    compare_pronunciation()
