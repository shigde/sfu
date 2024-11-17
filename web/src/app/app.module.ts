import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';

import {AppRoutingModule} from './app-routing.module';
import {AppComponent} from './app.component';
import {DashboardComponent} from './pages/dashboard/dashboard.component';
import {provideHttpClient, withInterceptorsFromDi} from '@angular/common/http';

import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {CommonModule} from '@angular/common';
import {httpInterceptorProviders, LoadingIndicatorComponent, ShigModule} from '@shigde/core';
import {LobbyEntryComponent} from './pages/lobby-entry/lobby-entry.component';
import {SettingsComponent} from './svg/settings.component';
import {SidebarComponent} from './component/sidebar/sidebar.component';
import {HeaderComponent} from './component/header/header.component';
import {ThumbnailCardComponent} from './component/thumbnail-card/thumbnail-card.component';


@NgModule({
  declarations: [
    AppComponent,
    DashboardComponent,
    LobbyEntryComponent,
    SettingsComponent
  ],
  bootstrap: [AppComponent], imports: [CommonModule,
    BrowserModule,
    AppRoutingModule,
    ReactiveFormsModule,
    FormsModule,
    ShigModule, LoadingIndicatorComponent, SidebarComponent, HeaderComponent, ThumbnailCardComponent], providers: [
    httpInterceptorProviders,
    provideHttpClient(withInterceptorsFromDi()),
  ]
})
export class AppModule {
  // ngDoBootstrap() {
  //     const customElement = createCustomElement(LobbyComponent, {injector: this.injector});
  //     customElements.define('shig-lobby', customElement);
  // }
}
