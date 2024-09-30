import {Component} from '@angular/core';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {SessionService} from '../../../../../../../shig-js-sdk/dist/core';

@Component({
  selector: 'app-password-forgotten',
  standalone: true,
  imports: [
    FormsModule,
    ReactiveFormsModule
  ],
  templateUrl: './password-forgotten.component.html',
  styleUrl: './password-forgotten.component.scss'
})
export class PasswordForgottenComponent {
  protected passForgetForm = new FormGroup({
    email: new FormControl('', [Validators.required, Validators.email]),
  });

  constructor(private router: Router, private session: SessionService) {
  }

  onSubmit() {
    if (this.passForgetForm.valid) {
      console.log(this.passForgetForm.value);
    }
  }
}
