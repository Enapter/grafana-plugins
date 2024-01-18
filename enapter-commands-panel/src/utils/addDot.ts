export const addDot = (s: string) => {
  if (!s.endsWith('.')) {
    return `${s}.`;
  }

  return s;
};
