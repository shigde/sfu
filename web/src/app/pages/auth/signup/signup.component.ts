import {Component} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {AuthService} from '@shigde/core';

@Component({
  selector: 'app-signup',
  standalone: true,
  imports: [
    ReactiveFormsModule
  ],
  templateUrl: './signup.component.html',
  styleUrl: './signup.component.scss'
})
export class SignupComponent {


  public signupForm = new FormGroup({
    user: new FormControl('', [Validators.required]),
    email: new FormControl('', [Validators.required, Validators.email]),
    password: new FormControl('', [Validators.required])
  });

  constructor(private readonly authService: AuthService, private readonly router: Router) {
  }

  public onSubmit() {
    if (this.signupForm.valid) {

      const user = `${this.signupForm.value.user}`;
      const email = `${this.signupForm.value.email}`;
      const password = `${this.signupForm.value.password}`;
      const account = {user, email, password};

      this.authService.registerAccount(account)
        .then((data:any) => {
          console.log(data);
          this.router.navigate(['/login']);
        })
        .catch((err: any) => console.log(err))
    }
  }
}
