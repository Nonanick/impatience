export const impatienceLSName = "ImpatienceCachedFiles";

export function readFromLS() {
  return JSON.parse(localStorage.getItem(impatienceLSName))
}

export function writeIntoLS(data) {
  (async() => localStorage.setItem(impatienceLSName, JSON.stringify(data)))();
}