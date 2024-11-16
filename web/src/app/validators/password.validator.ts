import {AbstractControl, ValidationErrors, ValidatorFn} from '@angular/forms';

export const PASSWORD_CONSTRAIN: RegExp = /^(?=[^A-Z]*[A-Z])(?=[^a-z]*[a-z])(?=\D*\d).{8,}$/;

const UPPERCASE_LETTER_CONSTRAIN: RegExp = /^(?=.*[A-Z])/;
const LOWERCASE_LETTER_CONSTRAIN: RegExp = /(?=.*[a-z])/;
const DIGIT_CONSTRAIN: RegExp = /(.*[0-9].*)/;
const SPECIAL_CHARACTER_CONSTRAIN: RegExp = /^(?=.*[!@#$%^&*/-_:;+`´,'"(){}≠|?])/;
const MIN_LENGTH_CONSTRAIN: RegExp = /.{8,}/;

export class PasswordValidator {

  public static confirm(oldPassword: string, newPassword: string): ValidatorFn {
    return (control: AbstractControl): ValidationErrors | null => {
      return control.get(oldPassword)?.value === control.get(newPassword)?.value ? null : {confirm: true};
    };
  }

  public static uppercaseLetter(control: AbstractControl): ValidationErrors | null {
    return control?.value?.match(UPPERCASE_LETTER_CONSTRAIN) ? null : {uppercaseLetter: true};
  }

  public static lowercaseLetter(control: AbstractControl): ValidationErrors | null {
    return control?.value?.match(LOWERCASE_LETTER_CONSTRAIN) ? null : {lowercaseLetter: true};
  }

  public static digit(control: AbstractControl): ValidationErrors | null {
    return control?.value?.match(DIGIT_CONSTRAIN) ? null : {digit: true};
  }

  public static specialCharacter(control: AbstractControl): ValidationErrors | null {
    return control?.value?.match(SPECIAL_CHARACTER_CONSTRAIN) ? null : {specialCharacter: true};
  }

  public static minLength(control: AbstractControl): ValidationErrors | null {
    return control?.value?.match(MIN_LENGTH_CONSTRAIN) ? null : {minLength: true};
  }
}
