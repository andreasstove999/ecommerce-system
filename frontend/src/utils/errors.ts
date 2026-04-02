export class AppError extends Error {
  status?: number;

  constructor(message: string, status?: number) {
    super(message);
    this.name = 'AppError';
    this.status = status;
  }
}

export const toErrorMessage = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message;
  }

  return 'Something went wrong. Please try again.';
};
