package catalog.service;

import catalog.domain.Product;
import catalog.repository.CatalogRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

class CatalogServiceTest {

    CatalogRepository repo;
    CatalogService service;

    @BeforeEach
    void setUp() {
        repo = mock(CatalogRepository.class);
        service = new CatalogService(repo);
    }

    @Test
    void listProducts_normalizesLimitAndOffset() {
        when(repo.list(50, 0)).thenReturn(List.of());

        service.listProducts(0, -10);

        verify(repo).list(50, 0);
    }

    @Test
    void getProduct_delegatesToRepo() {
        var id = UUID.randomUUID();
        when(repo.getById(id)).thenReturn(Optional.empty());

        var res = service.getProduct(id);

        assertThat(res).isEmpty();
        verify(repo).getById(id);
    }

    @Test
    void createProduct_delegatesToRepo() {
        var p = new Product();
        var created = new Product();
        when(repo.create(p)).thenReturn(created);

        var res = service.createProduct(p);

        assertThat(res).isSameAs(created);
        verify(repo).create(p);
    }
}
