import {Component} from '@angular/core';
import {FormControl, FormGroup, NgForm, ReactiveFormsModule, Validators} from '@angular/forms';
import {Router} from '@angular/router';
import {User, SessionService} from '@shigde/core';

@Component({
  selector: 'app-login',
  standalone: true,
  templateUrl: './login.component.html',
  imports: [
    ReactiveFormsModule
  ],
  styleUrls: ['./login.component.scss']
})
export class LoginComponent {
  private readonly authService: any;

  users: User[];
  protected loginForm = new FormGroup({
    email: new FormControl('', [Validators.required, Validators.email]),
    password: new FormControl('', [Validators.required])
  })

  constructor(private router: Router, private session: SessionService) {
    this.users = session.getUsers();
  }

  onSubmit(){
    if(this.loginForm.valid){
      console.log(this.loginForm.value);
      this.authService.login(this.loginForm.value)
        .subscribe((data: any) => {
          if(this.authService.isLoggedIn()){
            this.router.navigate(['/admin']);
          }
          console.log(data);
        });
    }
  }

  onLogin(f: NgForm): void {
    if (!f.value.user) {
      return;
    }
    if (this.session.setUserName(f.value.user)) {
      this.router.navigate(['']);
    }
  }
}
