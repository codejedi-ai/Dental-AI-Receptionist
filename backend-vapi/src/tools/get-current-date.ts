export function getCurrentDate(): string {
  const now = new Date();
  const date = now.toLocaleDateString("en-CA", {
    timeZone: "America/Toronto",
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  });
  const time = now.toLocaleTimeString("en-CA", {
    timeZone: "America/Toronto",
    hour: "2-digit",
    minute: "2-digit",
  });
  return `Today is ${date}, ${time} Toronto time.`;
}
