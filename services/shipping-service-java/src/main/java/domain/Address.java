package main.java.domain;

import jakarta.persistence.Embeddable;

@Embeddable
public class Address {
    public String line1;
    public String line2;
    public String city;
    public String state;
    public String postalCode;
    public String country;
}
