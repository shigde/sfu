import { ComponentFixture, TestBed } from '@angular/core/testing';

import { LobbyEntryComponent } from './lobby-entry.component';

describe('LobbyEntryComponent', () => {
  let component: LobbyEntryComponent;
  let fixture: ComponentFixture<LobbyEntryComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [LobbyEntryComponent]
    });
    fixture = TestBed.createComponent(LobbyEntryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
