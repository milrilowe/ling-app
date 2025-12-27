export type PhonemeCategory =
  | 'vowel'
  | 'diphthong'
  | 'stop'
  | 'fricative'
  | 'affricate'
  | 'nasal'
  | 'liquid'
  | 'glide'

export interface CanonicalPhoneme {
  ipa: string
  example: string
  highlight: [number, number] // [start, end] indices of letters to bold
  category: PhonemeCategory
}

// All 39 English phonemes from the pronunciation analysis system
export const ENGLISH_PHONEMES: CanonicalPhoneme[] = [
  // Vowels (10)
  { ipa: 'ɑ', example: 'father', highlight: [1, 2], category: 'vowel' }, // fAther
  { ipa: 'æ', example: 'cat', highlight: [1, 2], category: 'vowel' }, // cAt
  { ipa: 'ʌ', example: 'but', highlight: [1, 2], category: 'vowel' }, // bUt
  { ipa: 'ɔ', example: 'thought', highlight: [2, 4], category: 'vowel' }, // thOUght
  { ipa: 'ɛ', example: 'bed', highlight: [1, 2], category: 'vowel' }, // bEd
  { ipa: 'ɝ', example: 'bird', highlight: [1, 3], category: 'vowel' }, // bIRd
  { ipa: 'ɪ', example: 'bit', highlight: [1, 2], category: 'vowel' }, // bIt
  { ipa: 'i', example: 'bee', highlight: [1, 3], category: 'vowel' }, // bEE
  { ipa: 'ʊ', example: 'book', highlight: [1, 3], category: 'vowel' }, // bOOk
  { ipa: 'u', example: 'blue', highlight: [2, 4], category: 'vowel' }, // blUE

  // Diphthongs (5)
  { ipa: 'aʊ', example: 'cow', highlight: [1, 3], category: 'diphthong' }, // cOW
  { ipa: 'aɪ', example: 'buy', highlight: [1, 3], category: 'diphthong' }, // bUY
  { ipa: 'eɪ', example: 'say', highlight: [1, 3], category: 'diphthong' }, // sAY
  { ipa: 'oʊ', example: 'go', highlight: [1, 2], category: 'diphthong' }, // gO
  { ipa: 'ɔɪ', example: 'boy', highlight: [1, 3], category: 'diphthong' }, // bOY

  // Stops (6)
  { ipa: 'b', example: 'boy', highlight: [0, 1], category: 'stop' }, // Boy
  { ipa: 'd', example: 'dog', highlight: [0, 1], category: 'stop' }, // Dog
  { ipa: 'g', example: 'go', highlight: [0, 1], category: 'stop' }, // Go
  { ipa: 'k', example: 'cat', highlight: [0, 1], category: 'stop' }, // Cat
  { ipa: 'p', example: 'pot', highlight: [0, 1], category: 'stop' }, // Pot
  { ipa: 't', example: 'top', highlight: [0, 1], category: 'stop' }, // Top

  // Fricatives (9)
  { ipa: 'ð', example: 'this', highlight: [0, 2], category: 'fricative' }, // THis
  { ipa: 'f', example: 'fish', highlight: [0, 1], category: 'fricative' }, // Fish
  { ipa: 's', example: 'sun', highlight: [0, 1], category: 'fricative' }, // Sun
  { ipa: 'ʃ', example: 'she', highlight: [0, 2], category: 'fricative' }, // SHe
  { ipa: 'θ', example: 'think', highlight: [0, 2], category: 'fricative' }, // THink
  { ipa: 'v', example: 'voice', highlight: [0, 1], category: 'fricative' }, // Voice
  { ipa: 'z', example: 'zebra', highlight: [0, 1], category: 'fricative' }, // Zebra
  { ipa: 'ʒ', example: 'measure', highlight: [3, 4], category: 'fricative' }, // meaSure
  { ipa: 'h', example: 'hello', highlight: [0, 1], category: 'fricative' }, // Hello

  // Affricates (2)
  { ipa: 'tʃ', example: 'church', highlight: [0, 2], category: 'affricate' }, // CHurch
  { ipa: 'dʒ', example: 'judge', highlight: [0, 1], category: 'affricate' }, // Judge

  // Nasals (3)
  { ipa: 'm', example: 'man', highlight: [0, 1], category: 'nasal' }, // Man
  { ipa: 'n', example: 'no', highlight: [0, 1], category: 'nasal' }, // No
  { ipa: 'ŋ', example: 'sing', highlight: [2, 3], category: 'nasal' }, // siN(g is silent)

  // Liquids (2)
  { ipa: 'l', example: 'love', highlight: [0, 1], category: 'liquid' }, // Love
  { ipa: 'ɹ', example: 'red', highlight: [0, 1], category: 'liquid' }, // Red

  // Glides (2)
  { ipa: 'w', example: 'water', highlight: [0, 1], category: 'glide' }, // Water
  { ipa: 'j', example: 'yes', highlight: [0, 1], category: 'glide' }, // Yes
]

export const CATEGORY_LABELS: Record<PhonemeCategory, string> = {
  vowel: 'Vowels',
  diphthong: 'Diphthongs',
  stop: 'Stops',
  fricative: 'Fricatives',
  affricate: 'Affricates',
  nasal: 'Nasals',
  liquid: 'Liquids',
  glide: 'Glides',
}

// Order for displaying categories
export const CATEGORY_ORDER: PhonemeCategory[] = [
  'vowel',
  'diphthong',
  'stop',
  'fricative',
  'affricate',
  'nasal',
  'liquid',
  'glide',
]

// Helper to render example with highlighted letters
export function renderExample(example: string, highlight: [number, number]): {
  before: string
  highlighted: string
  after: string
} {
  return {
    before: example.slice(0, highlight[0]),
    highlighted: example.slice(highlight[0], highlight[1]),
    after: example.slice(highlight[1]),
  }
}
