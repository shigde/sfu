import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';

import {AppRoutingModule} from './app-routing.module';
import {AppComponent} from './app.component';
import {DashboardComponent} from './dashboard/dashboard.component';
import {HttpClientModule} from '@angular/common/http';

import {LiveStreamComponent} from './live-stream/live-stream.component';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {LoginComponent} from './login/login.component';
import {CommonModule} from '@angular/common';
import {httpInterceptorProviders, ShigModule} from '@shigde/core';
import {LobbyEntryComponent} from './lobby-entry/lobby-entry.component';
import {SettingsComponent} from './svg/settings.component';


@NgModule({
    declarations: [
        AppComponent,
        DashboardComponent,
        LiveStreamComponent,
        LoginComponent,
        LobbyEntryComponent,
        SettingsComponent
    ],
    imports: [
        CommonModule,
        BrowserModule,
        HttpClientModule,
        AppRoutingModule,
        ReactiveFormsModule,
        FormsModule,
        ShigModule
    ],
    providers: [
        httpInterceptorProviders,
    ],
    bootstrap: [AppComponent]
})
export class AppModule {

    // ngDoBootstrap() {
    //     const customElement = createCustomElement(LobbyComponent, {injector: this.injector});
    //     customElements.define('shig-lobby', customElement);
    // }
}
