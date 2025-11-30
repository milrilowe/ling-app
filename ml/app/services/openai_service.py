from openai import OpenAI
from app.config import settings


class OpenAIService:
    """Service for OpenAI API interactions."""

    def __init__(self):
        self.client = OpenAI(api_key=settings.openai_api_key) if settings.openai_api_key else None

    async def generate_conversation_prompt(self, difficulty: str = "beginner") -> str:
        """
        Generate a conversation prompt for the user.

        Args:
            difficulty: Difficulty level (beginner, intermediate, advanced)

        Returns:
            Generated conversation prompt
        """
        if not self.client:
            raise ValueError("OpenAI API key not configured")

        # TODO: Implement conversation generation
        # This is a placeholder
        response = self.client.chat.completions.create(
            model="gpt-4o",
            messages=[
                {"role": "system", "content": "You are a language learning assistant."},
                {"role": "user", "content": f"Generate a simple {difficulty} level English conversation prompt."}
            ]
        )
        return response.choices[0].message.content

    async def evaluate_response(self, user_text: str, expected_context: str) -> dict:
        """
        Evaluate user's response for content quality.

        Args:
            user_text: What the user said
            expected_context: Context of the conversation

        Returns:
            Evaluation results
        """
        if not self.client:
            raise ValueError("OpenAI API key not configured")

        # TODO: Implement response evaluation
        # This is a placeholder
        return {"score": 0.8, "feedback": "Good response!"}


# Singleton instance
openai_service = OpenAIService()
