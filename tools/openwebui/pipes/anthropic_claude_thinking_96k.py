"""
title: Anthropic Claude Thinking 96K
authors: lavantien, based on work by justinh-rahb and christian-taillon
author_url: https://github.com/lavantien
repo_url: https://github.com/lavantien/llm-tournament/tree/main/tools/openwebui/pipes/
version: 0.3
required_open_webui_version: 0.5.17+
license: MIT
"""

import os
import requests
import json
import time
from typing import List, Union, Generator, Iterator
from pydantic import BaseModel, Field
from open_webui.utils.misc import pop_system_message


class Pipe:
    class Valves(BaseModel):
        ANTHROPIC_API_KEY: str = Field(default="")

    def __init__(self):
        self.type = "manifold"
        self.id = "anthropic"
        self.name = "anthropic/"
        self.valves = self.Valves(
            **{"ANTHROPIC_API_KEY": os.getenv("ANTHROPIC_API_KEY", "")}
        )
        self.MAX_IMAGE_SIZE = 5 * 1024 * 1024  # 5MB per image
        self.DEFAULT_MAX_TOKENS = 128000
        self.DEFAULT_BUDGET_TOKENS = 96000
        pass

    def get_anthropic_models(self):
        return [
            {"id": "claude-3-opus-latest", "name": "claude-3-opus"},
            {"id": "claude-3-5-haiku-latest", "name": "claude-3.5-haiku"},
            {"id": "claude-3-5-sonnet-latest", "name": "claude-3.5-sonnet"},
            {"id": "claude-3-7-sonnet-latest", "name": "claude-3.7-sonnet"},
            {
                "id": "claude-3-7-sonnet-latest-thinking",
                "name": "claude-3.7-sonnet (thinking)",
            },
        ]

    def pipes(self) -> List[dict]:
        return self.get_anthropic_models()

    def process_image(self, image_data):
        """Process image data with size validation."""
        if image_data["image_url"]["url"].startswith("data:image"):
            mime_type, base64_data = image_data["image_url"]["url"].split(
                ",", 1)
            media_type = mime_type.split(":")[1].split(";")[0]

            # Check base64 image size
            # Convert base64 size to bytes
            image_size = len(base64_data) * 3 / 4
            if image_size > self.MAX_IMAGE_SIZE:
                raise ValueError(
                    f"Image size exceeds 5MB limit: {
                        image_size / (1024 * 1024):.2f}MB"
                )

            return {
                "type": "image",
                "source": {
                    "type": "base64",
                    "media_type": media_type,
                    "data": base64_data,
                },
            }
        else:
            # For URL images, perform size check after fetching
            url = image_data["image_url"]["url"]
            response = requests.head(url, allow_redirects=True)
            content_length = int(response.headers.get("content-length", 0))

            if content_length > self.MAX_IMAGE_SIZE:
                raise ValueError(
                    f"Image at URL exceeds 5MB limit: {
                        content_length / (1024 * 1024):.2f}MB"
                )

            return {
                "type": "image",
                "source": {"type": "url", "url": url},
            }

    def pipe(self, body: dict) -> Union[str, Generator, Iterator]:
        system_message, messages = pop_system_message(body["messages"])

        processed_messages = []
        total_image_size = 0

        for message in messages:
            processed_content = []
            if isinstance(message.get("content"), list):
                for item in message["content"]:
                    if item["type"] == "text":
                        processed_content.append(
                            {"type": "text", "text": item["text"]})
                    elif item["type"] == "image_url":
                        processed_image = self.process_image(item)
                        processed_content.append(processed_image)

                        # Track total size for base64 images
                        if processed_image["source"]["type"] == "base64":
                            image_size = len(
                                processed_image["source"]["data"]) * 3 / 4
                            total_image_size += image_size
                            if (
                                total_image_size > 100 * 1024 * 1024
                            ):  # 100MB total limit
                                raise ValueError(
                                    "Total size of images exceeds 100 MB limit"
                                )
            else:
                processed_content = [
                    {"type": "text", "text": message.get("content", "")}
                ]

            processed_messages.append(
                {"role": message["role"], "content": processed_content}
            )

        # Determine if using thinking mode
        model_name = body["model"][body["model"].find(".") + 1:]
        is_thinking_mode = "-thinking" in body["model"]
        if is_thinking_mode:
            # Strip the "-thinking" suffix for API call
            model_name = model_name.replace("-thinking", "")

        payload = {
            "model": model_name,
            "messages": processed_messages,
            "max_tokens": body.get("max_tokens", self.DEFAULT_MAX_TOKENS),
            # "temperature": body.get("temperature", 0.8),
            # "top_k": body.get("top_k", 40),
            # "top_p": body.get("top_p", 0.9),
            "stop_sequences": body.get("stop", []),
            **({"system": str(system_message)} if system_message else {}),
            "stream": body.get("stream", True),
        }

        # Add thinking parameters if using thinking mode
        if is_thinking_mode:
            payload["thinking"] = {
                "type": "enabled",
                "budget_tokens": body.get("budget_tokens", self.DEFAULT_BUDGET_TOKENS),
            }

        headers = {
            "x-api-key": self.valves.ANTHROPIC_API_KEY,
            "anthropic-version": "2023-06-01",
            "content-type": "application/json",
        }

        # Add additional headers for 128k output and prompt caching
        headers["anthropic-beta"] = "output-128k-2025-02-19"
        headers["prompt-caching-2024-07-31"] = ""

        url = "https://api.anthropic.com/v1/messages"

        try:
            if body.get("stream", True):
                return self.stream_response(url, headers, payload, is_thinking_mode)
            else:
                return self.non_stream_response(url, headers, payload, is_thinking_mode)
        except requests.exceptions.RequestException as e:
            print(f"Request failed: {e}")
            return f"Error: Request failed: {e}"
        except Exception as e:
            print(f"Error in pipe method: {e}")
            return f"Error: {e}"

    def stream_response(self, url, headers, payload, is_thinking_mode):
        try:
            with requests.post(
                url, headers=headers, json=payload, stream=True, timeout=(3.05, 60)
            ) as response:
                if response.status_code != 200:
                    raise Exception(
                        f"HTTP Error {response.status_code}: {response.text}"
                    )

                # Track if we're currently in thinking mode output to format appropriately
                in_thinking_section = False
                thinking_text = ""

                for line in response.iter_lines():
                    if line:
                        line = line.decode("utf-8")
                        if line.startswith("data: "):
                            try:
                                data = json.loads(line[6:])

                                # Handle thinking content
                                if (
                                    is_thinking_mode
                                    and data.get("type") == "content_block_start"
                                    and data.get("content_block", {}).get("type")
                                    == "thinking"
                                ):
                                    in_thinking_section = True
                                    thinking_prefix = "\n💭 [Thinking]\n\n"
                                    yield thinking_prefix
                                    if data.get("content_block", {}).get("thinking"):
                                        yield data["content_block"]["thinking"]

                                # Handle thinking delta
                                elif (
                                    is_thinking_mode
                                    and data.get("type") == "content_block_delta"
                                    and in_thinking_section
                                ):
                                    if data.get("delta", {}).get("thinking"):
                                        yield data["delta"]["thinking"]

                                # Handle thinking end
                                elif (
                                    is_thinking_mode
                                    and data.get("type") == "content_block_stop"
                                    and in_thinking_section
                                ):
                                    in_thinking_section = False

                                # Handle regular text content
                                elif (
                                    data.get("type") == "content_block_start"
                                    and data.get("content_block", {}).get("type")
                                    == "text"
                                ):
                                    if (
                                        is_thinking_mode
                                        and not in_thinking_section
                                        and not thinking_text
                                    ):
                                        yield "\n\n🔍 [Response]\n\n"
                                    yield data["content_block"]["text"]

                                # Handle regular text delta
                                elif (
                                    data.get("type") == "content_block_delta"
                                    and not in_thinking_section
                                ):
                                    if data.get("delta", {}).get("text"):
                                        yield data["delta"]["text"]

                                # Handle message stop
                                elif data.get("type") == "message_stop":
                                    break

                                # Handle full message (for non-stream fallback)
                                elif data.get("type") == "message":
                                    # Process thinking content first if available
                                    if is_thinking_mode:
                                        for content in data.get("content", []):
                                            if content.get("type") == "thinking":
                                                yield "\n💭 [Thinking] " + content.get(
                                                    "thinking", ""
                                                )
                                                yield "\n\n🔍 [Response] "
                                                break

                                    # Then process text content
                                    for content in data.get("content", []):
                                        if content.get("type") == "text":
                                            yield content.get("text", "")

                                time.sleep(
                                    0.01
                                )  # Delay to avoid overwhelming the client

                            except json.JSONDecodeError:
                                print(f"Failed to parse JSON: {line}")
                            except KeyError as e:
                                print(f"Unexpected data structure: {e}")
                                print(f"Full data: {data}")
        except requests.exceptions.RequestException as e:
            print(f"Request failed: {e}")
            yield f"Error: Request failed: {e}"
        except Exception as e:
            print(f"General error in stream_response method: {e}")
            yield f"Error: {e}"

    def non_stream_response(self, url, headers, payload, is_thinking_mode):
        try:
            response = requests.post(
                url, headers=headers, json=payload, timeout=(3.05, 60)
            )
            if response.status_code != 200:
                raise Exception(
                    f"HTTP Error {response.status_code}: {response.text}")

            res = response.json()

            if is_thinking_mode:
                result = ""
                # Check for thinking content
                for content in res.get("content", []):
                    if content.get("type") == "thinking":
                        result += f"\n💭 [Thinking] {
                            content.get('thinking', '')}\n\n"
                    elif content.get("type") == "text":
                        result += f"🔍 [Response] {content.get('text', '')}"
                return result
            else:
                # Normal response processing
                return (
                    res["content"][0]["text"]
                    if "content" in res and res["content"]
                    else ""
                )
        except requests.exceptions.RequestException as e:
            print(f"Failed non-stream request: {e}")
            return f"Error: {e}"
