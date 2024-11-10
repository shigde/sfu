import {Component} from '@angular/core';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute, Router} from '@angular/router';
import {AuthService} from '@shigde/core';
import {NgIf} from '@angular/common';
import {catchError, of, take, tap} from 'rxjs';
import {PASSWORD_CONSTRAIN, PasswordValidator} from '../../../validators/password.validator';

@Component({
  selector: 'app-update-password',
  standalone: true,
  imports: [
    FormsModule,
    ReactiveFormsModule,
    NgIf
  ],
  templateUrl: './update-password.component.html',
  styleUrl: './update-password.component.scss'
})
export class UpdatePasswordComponent {

  public fail = false;
  public success = false;
  private token: string | undefined;

  protected updateForm = new FormGroup({
    orgPassword: new FormControl('', [Validators.required]),
    newPassword: new FormControl('', [
      Validators.required,
      Validators.pattern(PASSWORD_CONSTRAIN),
      PasswordValidator.uppercaseLetter,
      PasswordValidator.lowercaseLetter,
      PasswordValidator.digit,
      PasswordValidator.specialCharacter,
      PasswordValidator.minLength
    ]),
    confirmPassword: new FormControl('', [Validators.required]),
  }, {validators: [PasswordValidator.confirm('confirmPassword', 'newPassword')]});

  constructor(private router: Router, private readonly route: ActivatedRoute, private authService: AuthService) {
  }

  ngOnInit(): void {
    const token = this.route.snapshot.paramMap.get('token');
    if (token === null) {
      this.fail = true;
      return;
    }

    this.token = token;
  }

  onSubmit() {
    this.success = false;
    if (this.updateForm.valid && this.token !== undefined) {
      const orgPass = `${this.updateForm.value.orgPassword}`;
      const newPass = `${this.updateForm.value.newPassword}`;
      this.authService.updatePassword(orgPass, newPass).pipe(
        take(1),
        tap(a => this.success = true),
        catchError((_) => this.handleError())
      ).subscribe();
    } else {
      this.updateForm.get('orgPassword')?.markAsTouched({onlySelf: true});
      this.updateForm.get('newPassword')?.markAsTouched({onlySelf: true});
      this.updateForm.get('confirmPassword')?.markAsTouched({onlySelf: true});
    }
  }

  private handleError() {
    this.fail = true;
    return of('');
  }
}
