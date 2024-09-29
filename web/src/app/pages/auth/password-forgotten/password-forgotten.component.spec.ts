import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PasswordForgottenComponent } from './password-forgotten.component';

describe('PasswordForgottenComponent', () => {
  let component: PasswordForgottenComponent;
  let fixture: ComponentFixture<PasswordForgottenComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PasswordForgottenComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PasswordForgottenComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
