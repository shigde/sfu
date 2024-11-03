import {Component, OnInit} from '@angular/core';
import {AuthService} from '@shigde/core';
import {catchError, of, take, tap} from 'rxjs';
import {ActivatedRoute} from '@angular/router';
import {NgIf} from '@angular/common';

@Component({
  selector: 'app-activate-account',
  standalone: true,
  imports: [
    NgIf
  ],
  templateUrl: './activate-account.component.html',
  styleUrl: './activate-account.component.scss'
})
export class ActivateAccountComponent implements OnInit{
  public success = false;
  public fail = false;

  constructor(
    private readonly authService: AuthService,
    private readonly route: ActivatedRoute
  ) {

  }

  ngOnInit(): void {
    const token = this.route.snapshot.paramMap.get('token');
    if (token === null) {
      this.fail = true;
      return;
    }

    this.authService.verifyAccount(token).pipe(
      take(1),
      tap(a => this.success = true),
      catchError((_) => this.handleError())
    ).subscribe();
  }

  private handleError() {
    this.fail = true;
    return of('');
  }
}
