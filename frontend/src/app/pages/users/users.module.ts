import { NgModule } from '@angular/core';
import { UserListComponent } from './user-list/user-list.component';
import {UsersRoutingModule} from "./users-routing.module";

@NgModule({
  declarations: [
    UserListComponent
  ],
  imports: [
    UsersRoutingModule
  ]
})
export class UsersModule { }
