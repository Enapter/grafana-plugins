import {
  Argument,
  BooleanArgument,
  FloatArgument,
  IntegerArgument,
  StringArgument,
} from '../types/types';

type ArgValidatorType = {
  isValueValid: () => boolean;
  getErrorMessage: () => string;
};

type ArgValidatorOptions = {
  skip?: {
    required?: boolean;
  };
};

abstract class BaseArgValidator<T extends Argument = Argument> implements ArgValidatorType {
  protected readonly arg: T;
  protected readonly options: ArgValidatorOptions;
  protected errorMessage: string | null = null;

  constructor(arg: T, options?: ArgValidatorOptions) {
    this.arg = arg;
    this.options = options || {};
  }

  public isValueValid(): boolean {
    throw new Error('Method not implemented.');
  }

  public getErrorMessage(): string {
    return this.errorMessage || '';
  }

  static isArgBoolean(arg: Argument): arg is BooleanArgument {
    return arg.type === 'boolean';
  }

  static isArgString(arg: Argument): arg is StringArgument {
    return arg.type === 'string';
  }

  static isArgInteger(arg: Argument): arg is IntegerArgument {
    return arg.type === 'integer';
  }

  static isArgFloat(arg: Argument): arg is FloatArgument {
    return arg.type === 'float';
  }

  protected isEmpty(): boolean {
    return this.value === undefined || this.value === '';
  }

  protected isRequired(): boolean {
    return !!this.arg.required && !this.shouldSkipRequired();
  }

  protected get value(): T['value'] {
    return this.arg.value;
  }

  protected shouldSkipRequired(): boolean {
    return !!this.options.skip?.required;
  }
}

class BooleanArgValidator extends BaseArgValidator<BooleanArgument> {
  isValueValid() {
    return true;
  }
}

class StringArgValidator extends BaseArgValidator<StringArgument> {
  isValueValid() {
    if (this.isRequired() && this.isEmpty()) {
      this.errorMessage = 'Argument is required';

      return false;
    }

    return true;
  }
}

class NumericArgValidator extends BaseArgValidator<FloatArgument | IntegerArgument> {
  private readonly coercedValue: number;

  constructor(arg: FloatArgument | IntegerArgument, options?: ArgValidatorOptions) {
    super(arg, options);
    this.coercedValue = Number(this.value);
  }

  isValueValid(): boolean {
    if (this.isRequired() && this.isEmpty()) {
      this.errorMessage = 'Argument is required';

      return false;
    }

    if (!this.isRequired() && this.isEmpty()) {
      return true;
    }

    if (this.isExtrapolated()) {
      return true;
    }

    if (!this.isNumeric()) {
      this.errorMessage = 'Argument should be a number or a variable';

      return false;
    }

    if (BaseArgValidator.isArgInteger(this.arg) && !this.isValidInteger()) {
      this.errorMessage = 'Argument should be an integer or a variable';

      return false;
    }

    if (BaseArgValidator.isArgFloat(this.arg) && !this.isValidFloat()) {
      this.errorMessage = 'Argument should be a float or a variable';

      return false;
    }

    return this.validateValueWithinRange();
  }

  private isExtrapolated(): boolean {
    return !!this.value?.startsWith('$');
  }

  private isNumeric(): boolean {
    return !Number.isNaN(this.coercedValue);
  }

  private isValidInteger(): boolean {
    return Number.isInteger(this.coercedValue);
  }

  private isValidFloat(): boolean {
    return this.isNumeric();
  }

  private validateValueWithinRange(): boolean {
    const min = this.arg.min;
    const max = this.arg.max;

    if (min !== undefined && this.coercedValue < min) {
      this.errorMessage = `Argument should be greater or equal to ${min}`;

      return false;
    }

    if (max !== undefined && this.coercedValue > max) {
      this.errorMessage = `Argument should be less or equal to ${max}`;

      return false;
    }

    return true;
  }
}

export class ArgValidator implements ArgValidatorType {
  private validator: ArgValidatorType;

  constructor(arg: Argument, options?: ArgValidatorOptions) {
    if (BaseArgValidator.isArgBoolean(arg)) {
      this.validator = new BooleanArgValidator(arg, options);
    } else if (BaseArgValidator.isArgString(arg)) {
      this.validator = new StringArgValidator(arg, options);
    } else if (BaseArgValidator.isArgInteger(arg) || BaseArgValidator.isArgFloat(arg)) {
      this.validator = new NumericArgValidator(arg, options);
    } else {
      throw new Error('Unsupported argument type');
    }
  }

  isValueValid(): boolean {
    return this.validator.isValueValid();
  }

  getErrorMessage(): string {
    return this.validator.getErrorMessage();
  }
}
